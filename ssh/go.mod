module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/gliderlabs/ssh v0.3.4
	github.com/go-resty/resty/v2 v2.7.0
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pires/go-proxyproto v0.6.2
	github.com/shellhub-io/shellhub v0.8.1
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b
	golang.org/x/net v0.0.0-20211209124913-491a49abca63
)

replace github.com/shellhub-io/shellhub => ../
