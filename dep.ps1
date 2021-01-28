# Get the depencencies for this Go application.
$Env:GOPATH=Get-Location
$cmd = "go get " +
    "github.com/Azure/azure-sdk-for-go/storage " +
    "github.com/aws/aws-sdk-go/aws " +
    "github.com/aws/aws-sdk-go/aws/awserr " +
    "github.com/aws/aws-sdk-go/aws/session " +
    "github.com/aws/aws-sdk-go/service/dynamodb " +
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute " +
    "github.com/aws/aws-sdk-go/service/dynamodb/expression " +
    "github.com/google/uuid"
Invoke-Expression $cmd