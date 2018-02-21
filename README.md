# terraform-provider-ibm-api
Added the helm chart for the Micro service that exposes REST API for IBM Cloud terraform provider plugin

* List of parameter which can be provided by user while installing the helm chart

|Variable Name|Description|Default Value|
|-------------|-----------|-------------|
|image.repository|Image for api service||
|image.pullPolicy|Image Pull Policy|IfNotPresent|
|pvc.enabled| whether to create pvc pr not|true|
|pvc.storageClass|pvc storage class|ibmc-file-bronze|
|pvc.mountPath|pvc mount path|/tmp|
|pvc.name|Name of pvc|tfstate-data|
|ingress.path|Ingress Path|/|
|ingress.host|ingress hostname||

## How to use this helm chart

* helm repo add terraformibmprovider-charts  https://ibmterraform.github.io/terraform-provider-ibm-api/

* To get the availale charts in the repo 
    helm search

* helm install <chartname>  --set param=value

for eg: helm install <chartname>  --set pvc.enabled=false --set ingress.host=<ingress_hostname>

 
