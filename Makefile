.PHONY: build

build:
	sam build

package:
	rm -f poc_receive_payment_lambda
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go && zip poc_receive_payment_lambda.zip bootstrap
