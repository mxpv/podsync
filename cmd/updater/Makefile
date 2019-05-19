TIMEOUT := 300
ROLE := AWSLambdaBasicExecutionRole
NAME := Updater

build:
	pip3 install --requirement requirements.txt --target package
	cd package && zip -r9 ../function.zip .
	zip -g function.zip function.py updater.py

deploy: build
	aws --profile Podsync lambda create-function \
		--function-name $(NAME) \
		--role $(shell aws --profile Podsync iam get-role --role-name $(ROLE) --query 'Role.Arn' --output text) \
		--runtime python3.7 \
		--handler function.handler \
		--zip-file fileb://function.zip \
		--timeout $(TIMEOUT) \
		--memory-size 128

update: build
	aws --profile Podsync lambda update-function-code \
		--function-name $(NAME) \
		--zip-file fileb://function.zip

.PHONY: push
push:
	docker build -t mxpv/updater .
	docker push mxpv/updater

clean:
	rm -rf package function.zip

.PHONY: deploy update clean
