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

KUBECONFIG="$(kind get kubeconfig-path --name kind)"

[ -f "${KUBECONFIG}" ] || die "Can't find kube-config at '${KUBECONFIG}'"

if ! helm init --debug --kubeconfig ${KUBECONFIG} ; then
    die "On bootstraping Helm"
fi
