ENV_OVERRIDE ?= ./.env.override

ifneq (,$(wildcard $(ENV_OVERRIDE)))
    include $(ENV_OVERRIDE)
endif

DOCKER_COMPOSE = ./bin/docker-compose
KEYGEN = ./bin/keygen

define COMPOSE_TEMPLATE
version: '3.7'
services:
  mongo:
    image: mongo:_VERSION_
endef

export COMPOSE_TEMPLATE

# Generate required private key for api service
api_private_key:
	@$(KEYGEN) genrsa -out api_private_key 2048

# Generate required public key for api service
api_public_key:
	@$(KEYGEN) rsa -in api_private_key -out api_public_key -pubout

# Generate required private key for ssh service
ssh_private_key:
	@$(KEYGEN) genrsa -out ssh_private_key 2048

.PHONY: keygen
# Generate required keys
keygen: api_private_key api_public_key ssh_private_key

.PHONY: start
## Start services
start: keygen
ifeq ($(SHELLHUB_ENV),development)
	@echo Starting ShellHub in development mode...
	@echo
	@$(DOCKER_COMPOSE) up
else
	@echo Starting ShellHub in production mode...
	@echo

	@$(DOCKER_COMPOSE) up -d
endif

.PHONY: stop
## Stop services
stop:
	@$(DOCKER_COMPOSE) stop

restart: stop start

.PHONY: build
## Build all services (append "SERVICE=<service>" to build a specific one)
build: check_development_mode
	@$(DOCKER_COMPOSE) build $(SERVICE)

.PHONY: check_development
# Check for development mode
check_development_mode:
ifneq ($(SHELLHUB_ENV),development)
	@echo Development mode disabled!
	@exit 1
endif

.SILENT:
.ONESHELL:
upgrade_mongodb: SHELL := sh
upgrade_mongodb:
	# Wait for mongodb to be available
	wait_for_mongo() {
		# Command used to ping mongodb
		MONGO_PING_CMD=$$(cat <<-EOF
			db.runCommand({ ping: 1 })
		EOF
		)

		while ! $(DOCKER_COMPOSE) exec mongo mongo --quiet --eval "quit($${MONGO_PING_CMD}.ok ? 0 : 1)" >/dev/null 2>&1; do
			sleep 1
		done
	}

	# Start mongodb using version provided by $1
	start_mongodb() {
		# Convert version to suitable format for 'setFeatureCompatibilityVersion'
		SERIES_VERSION=$$(echo $$1 | sed 's,\.[0-9]\+$$,,g')
		MAJOR_VERSION=$$(echo $$1 | cut -d. -f1)

		export EXTRA_COMPOSE_FILE=$$(mktemp)
		echo "$$COMPOSE_TEMPLATE" | sed "s,_VERSION_,$$1,g" > $$EXTRA_COMPOSE_FILE

		$(DOCKER_COMPOSE) stop mongo
		$(DOCKER_COMPOSE) up -d mongo

		# Command used to set compatibility version
		MONGO_SET_COMPAT_VERSION_CMD=$$(cat <<-EOF
			db.adminCommand({
				setFeatureCompatibilityVersion: '$$SERIES_VERSION'
			})
		EOF
		)

		# The legacy mongo shell is removed from MongoDB 6.0. The replacement is mongosh
		if [ $$MAJOR_VERSION -ge 6 ]; then
			$(DOCKER_COMPOSE) exec mongo ln -sf /usr/bin/mongosh /usr/bin/mongo
		fi

		if [ $$MAJOR_VERSION -ge 7 ]; then
			MONGO_SET_COMPAT_VERSION_CMD=$$(cat <<-EOF
				db.adminCommand({
					setFeatureCompatibilityVersion: '$$SERIES_VERSION',
					confirm: true
				})
			EOF
			)
		fi

		wait_for_mongo

		$(DOCKER_COMPOSE) exec mongo mongo --quiet --eval "quit($${MONGO_SET_COMPAT_VERSION_CMD}.ok ? 0 : 1)"
	}

	MONGO_CONTAINER_ID=$$($(DOCKER_COMPOSE) images -q mongo)

	[ -z "$$MONGO_CONTAINER_ID" ] && exit 0

	CURRENT_MONGO_VERSION=$$(docker image inspect \
		--format '{{range .RepoTags}}{{.}} {{end}}' \
		$$MONGO_CONTAINER_ID | tr -d ' ' | rev | cut -d' ' -f1 | rev | cut -d':' -f2
	)

	test "$${CURRENT_MONGO_VERSION#*4.2*}" != "$$CURRENT_MONGO_VERSION" && exit 0
	echo "Upgrading MongoDB instance..."

	start_mongodb $$CURRENT_MONGO_VERSION

	# Command used to get compatibility version
	MONGO_GET_COMPAT_VERSION_CMD=$$(cat <<-EOF
		db.adminCommand({
			getParameter: 1,
			featureCompatibilityVersion: 1
		})['featureCompatibilityVersion']
	EOF
	)

	# Before upgrading to 4.2-series, we need to upgrade earlier versions first
	while true; do
		# Starting from 4.0-series the 'featureCompatibilityVersion' is a object with a 'version' key
		MONGO_COMPAT_VERSION=$$($(DOCKER_COMPOSE) exec mongo mongo \
			--quiet \
			--eval "$${MONGO_GET_COMPAT_VERSION_CMD}.version" | tr -d '\r'
		)

		# Fallback to 3.0-series where the 'featureCompatibilityVersion' is a integer with the version
		if [ -z "$$MONGO_COMPAT_VERSION" ]; then
			MONGO_COMPAT_VERSION=$$($(DOCKER_COMPOSE) exec mongo mongo \
				--quiet \
				--eval "$$MONGO_GET_COMPAT_VERSION_CMD" | tr -d '\r'
			)
		fi

		case $$MONGO_COMPAT_VERSION in
			3.4*)
				echo "Upgrading MongoDB from 3.4-series to 3.6-series..."
				start_mongodb 3.6.21
				;;
			3.6*)
				echo "Upgrading MongoDB from 3.6-series to 4.0-series..."
				start_mongodb 4.0.22
				;;
			4.0*)
				echo "Upgrading MongoDB from 4.0-series to 4.2-series..."
				start_mongodb 4.2.12
				;;
			4.2*)
				echo "Upgrading MongoDB from 4.2-series to 4.4-series..."
				start_mongodb 4.4.8
				;;
			4.4*)
				echo "Upgrading MongoDB from 4.4-series to 5.0-series..."
				start_mongodb 5.0.6
				;;
			5.0*)
			 	echo "Upgrading MongoDB from 5.0-series to 6.0-series..."
			 	start_mongodb 6.0.2
			 	;;
			6.0*)
			 	echo "Upgrading MongoDB from 6.0-series to 7.0-series..."
			 	start_mongodb 7.0.2
			 	;;
			7.0*)
				echo "MongoDB upgrade successful!"
				break
				;;
	      esac
	done

	$(DOCKER_COMPOSE) stop mongo

.PHONY: help
help:
	@cat <<-EOF
		Available commands:
		
		build      Build all services (append "SERVICE=<service>" to build a specific one)
		start      Start services
		stop       Stop services
		EOF

.DEFAULT_GOAL := help
