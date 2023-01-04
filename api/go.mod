module github.com/shellhub-io/shellhub/api

go 1.14

require (
	github.com/cnf/structhash v0.0.0-20201127153200-e1b16c1ebc08
	github.com/emirpasic/gods v1.18.1
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/getsentry/sentry-go v0.13.0
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/hibiken/asynq v0.23.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v4 v4.10.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/shellhub-io/shellhub v0.5.2
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/square/mongo-lock v0.0.0-20201208161834-4db518ed7fb2
	github.com/stretchr/testify v1.8.1
	github.com/undefinedlabs/go-mpatch v1.0.6
	github.com/xakep666/mongo-migrate v0.2.1
	go.mongodb.org/mongo-driver v1.11.1
	golang.org/x/crypto v0.4.0
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
)

replace github.com/shellhub-io/shellhub => ../
