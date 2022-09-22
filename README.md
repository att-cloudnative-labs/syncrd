## Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to
Helm's [documentation](https://helm.sh/docs) to get started.

Once Helm has been set up correctly, add the repo as follows:

  helm repo add att-cloudnative-labs https://att-cloudnative-labs.github.io/syncrd

If you had already added this repo earlier, run `helm repo update` to retrieve
the latest versions of the packages.  You can then run `helm search repo
att-cloudnative-labs` to see the charts.

To install the syncrd chart:

    helm install syncrd att-cloudnative-labs/syncrd

To uninstall the chart:

    helm delete syncrd