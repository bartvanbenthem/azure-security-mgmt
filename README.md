# Description
Query all the Windows and Linux virtual machines within a specific subscription and the update management status from the log analytics workspace extension.

# Description
Module that contains CLI tools and the supporting packages for Azure development and administration. (Azure-sdk-for-go is used)

## prerequisites
Install azure cli: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest

## create azure spn

#### set variables for creating app registration
``` shell
$ spname='<<name-spn>>'
$ tenantId=$(az account show --query tenantId -o tsv)
$ subscriptions=('<<subscription-id-01 subscription-id-02 ...>>')
```
    
#### Create the Azure AD application
``` shell
$ applicationId=$(az ad app create \
    --display-name "$spname" \
    --identifier-uris "https://$spname" \
    --query appId -o tsv)
```

#### Update the application group memebership claims
``` shell
$ az ad app update --id $applicationId --set groupMembershipClaims=All
```

#### Create a service principal for the Azure AD application
``` shell
$ az ad sp create --id $applicationId
```

#### Get the service principal secret
``` shell
$ applicationSecret=$(az ad sp credential reset \
    --name $applicationId \
    --credential-description "passwrd" \
    --query password -o tsv)
```

#### Add SPN to the subscriptions as an reader
``` shell
for s in "${subscriptions[@]}"; do {
    az role assignment create --assignee $applicationId --subscription $s --role 'Reader'
}; done
```

## set environment variables for auth
Once the Azure App registration is created set the following environment variables:
``` shell

$ export AZURE_CLIENT_ID='$applicationId'
$ export AZURE_TENANT_ID=$tenantId
$ export AZURE_CLIENT_SECRET='$applicationSecret'
```

## install (Linux)
``` shell
$ git clone https://github.com/bartvanbenthem/azure-update-mgmt.git
$ sudo cp azure-update-mgmt/tree/master/clitools/bin/* /usr/bin/
```

## azm-win-updatemgmt
Tool to query all the Windows virtual machines within a specific subscription and the update management status from the log analytics workspace extension.

#### set environment variables and run 
``` shell
$ export AZURE_SUBSCRIPTION_ID='<<subscription-id>>'
$ export AZURE_MANAGED_BY_TAGGING_KEY='<<managedby-key-name>>'
$ export AZURE_MANAGED_BY_TAGGING_VALUE='<<managedby-value-name>>'
$ export AZURE_TENANT_NAME='<<tenant-name>>'
$ azure-update-mgmt
```