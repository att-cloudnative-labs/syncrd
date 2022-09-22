```shell
helm package syncrd
aws ecr get-login-password --region us-west-2 | helm registry login --username AWS --password-stdin 367800898930.dkr.ecr.us-west-2.amazonaws.com
helm push syncrd-helm-0.0.2.tgz oci://367800898930.dkr.ecr.us-west-2.amazonaws.com/
```