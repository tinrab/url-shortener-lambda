set GOOS=linux

go build -o shorten main.go
build-lambda-zip -o deployment.zip shorten

aws lambda create-function --region us-east-1 --function-name ShortenFunction --zip-file fileb://./deployment.zip --runtime go1.x --tracing-config Mode=Active --role arn:aws:iam::766594786016:role/lambda_basic_execution --handler shorten
