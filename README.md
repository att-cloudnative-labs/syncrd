# Syncrd

---

<p align="center">
  <a href="https://goreportcard.com/report/github.com/att-cloudnative-labs/syncrd" alt="Go Report Card">
    <img src="https://goreportcard.com/badge/github.com/att-cloudnative-labs/syncrd">
  </a>	
</p>
<p align="center">
    <a href="https://github.com/att-cloudnative-labs/syncrd/graphs/contributors" alt="Contributors">
		<img src="https://img.shields.io/github/contributors/att-cloudnative-labs/syncrd.svg">
	</a>
	<a href="https://github.com/att-cloudnative-labs/syncrd/commits/master" alt="Commits">
		<img src="https://img.shields.io/github/commit-activity/m/att-cloudnative-labs/syncrd.svg">
	</a>
	<a href="https://github.com/att-cloudnative-labs/syncrd/pulls" alt="Open pull requests">
		<img src="https://img.shields.io/github/issues-pr-raw/att-cloudnative-labs/syncrd.svg">
	</a>
	<a href="https://github.com/att-cloudnative-labs/syncrd/pulls" alt="Closed pull requests">
    	<img src="https://img.shields.io/github/issues-pr-closed-raw/att-cloudnative-labs/syncrd.svg">
	</a>
	<a href="https://github.com/att-cloudnative-labs/syncrd/issues" alt="Issues">
		<img src="https://img.shields.io/github/issues-raw/att-cloudnative-labs/syncrd.svg">
	</a>
	</p>
<p align="center">
	<a href="https://github.com/att-cloudnative-labs/syncrd/stargazers" alt="Stars">
		<img src="https://img.shields.io/github/stars/att-cloudnative-labs/syncrd.svg?style=social">
	</a>	
	<a href="https://github.com/att-cloudnative-labs/syncrd/watchers" alt="Watchers">
		<img src="https://img.shields.io/github/watchers/att-cloudnative-labs/syncrd.svg?style=social">
	</a>	
	<a href="https://github.com/att-cloudnative-labs/syncrd/network/members" alt="Forks">
		<img src="https://img.shields.io/github/forks/att-cloudnative-labs/syncrd.svg?style=social">
	</a>	
</p>

----
Syncrd is a custom-controller that allows automatic synchronization of any namespaced kubernetes resource, including custom resources. This controller includes a custom resource, "Syncr", that indicates that group/type of the resource to be copied along with the namespace and name. The resource also has a field for matchLabels to restrict which namespaces the resource is copied to.

The following fields of the Syncr spec indicate the source resource to be copied:
* apiGroup (for core resources, use an empty string)
* apiVersion
* kind
* namespace
* name

The following field is used to restrict the namespaces the resource is copied to:
* matchLabels

Here is a sample Syncr definition for syncing a core resource:
```$yaml
apiVersion: syncrd.atteg.com/v1beta1
kind: Syncr
metadata:
  name: samplesyncr
  namespace: default
spec:
  apiGroup: ""
  apiResource: kconfigs
  apiVersion: v1alpha1
  matchLabels:
    mylabel: myvalue
  name: sourceresourcename
  namespace: default
```

Here is a sample Syncr definition for syncing a custom resource:
```$yaml
apiVersion: syncrd.atteg.com/v1beta1
kind: Syncr
metadata:
  name: samplesyncr
  namespace: default
spec:
  apiGroup: customgroup.atteg.com
  apiResource: customKind
  apiVersion: v1beta1
  matchLabels:
    somelabel: somevalue
  name: resourcename
  namespace: default
```