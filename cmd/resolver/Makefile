
build:
	pip3 install --requirement requirements.txt --target package
	cd package && zip -r9 ../function.zip .
	zip -g function.zip lambda.py resolver.py

deploy: build
	aws --profile Podsync lambda create-function \
		--function-name Resolver \
		--role $(shell aws --profile Podsync iam get-role --role-name PodsyncResolverLambdaRole --query 'Role.Arn' --output text) \
		--runtime python3.7 \
		--handler lambda.handler \
		--zip-file fileb://function.zip \
		--timeout 10 \
		--memory-size 128

update: build
	aws --profile Podsync lambda update-function-code \
		--function-name Resolver \
		--zip-file fileb://function.zip

.PHONY: push
push:
	docker build -t mxpv/resolver .
	docker push mxpv/resolver


clean:
	rm -rf package function.zip

.PHONY: deploy update clean
