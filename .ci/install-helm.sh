#!/bin/bash

HELM_VERSION="v2.11.0"
HELM_URL="https://storage.googleapis.com"
HELM_URL_PATH="kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz"
HELM_TARGET_DIR="${HELM_TARGET_DIR:-/home/travis/bin}"
HELM_BIN="${HELM_TARGET_DIR}/helm"


function die () {
    echo "[ERROR] ${*}" 1>&2
    exit 1
}

# using kind configuration
KUBECONFIG="$(kind get kubeconfig-path --name kind)"
[ -f "${KUBECONFIG}" ] || die "Can't find kube-config at '${KUBECONFIG}'"

function kind_kubectl () {
    kubectl --kubeconfig ${KUBECONFIG} --namespace kube-system $*
}

# downloading helm
curl --location --output helm.tar.gz ${HELM_URL}/${HELM_URL_PATH} || die "On downloading Helm"

# unpacking tarball
tar zxvpf helm.tar.gz || die "On unpacking tarball"

# moving to final location
mv -v ./linux-amd64/helm ${HELM_BIN} || die "On moving helm-bin to final path"

# setting execution flag
chmod +x ${HELM_BIN} || die "On adding execution permission to helm-bin"

# cleaning up
rm -rfv ./linux-amd64 helm.tar.gz > /dev/null 2>&1

# giving cluster permissions to helm
kind_kubectl create serviceaccount tiller || die "On creating tiller service-account"
kind_kubectl create clusterrolebinding tiller-binding \
    --clusterrole cluster-admin \
    --serviceaccount kube-system:tiller || \
        die "On creating cluster-role-binding"

if ! helm init --debug --kubeconfig ${KUBECONFIG} --service-account tiller ; then
    die "On bootstraping Helm"
fi
