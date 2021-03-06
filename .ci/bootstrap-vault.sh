#/bin/bash

export VAULT_TOKEN="vault-root-token"
export VAULT_ADDR="http://127.0.0.1:8200"

function die () {
    echo "[ERROR] ${*}" 1>&2
    exit 1
}

function wait_for_vault () {
    max_attemtps=10
    attempts=0

    echo "Waiting for Vault at '${VAULT_ADDR}'..."
    until curl --fail ${VAULT_ADDR} > /dev/null 2>&1 || [ $attempts -eq $max_attemtps ] ; do
        echo "# Failed to reach Vault at '${VAULT_ADDR}' (${attempts}/${max_attemtps})"
        sleep $(( attempts++ ))
    done

    if [ $attempts -eq $max_attemtps ]; then
        die "Can't reach Vault at '${VAULT_ADDR}', timeout!"
    fi
}

function enable_secrets_kv () {
    if ! vault read sys/mounts |grep -q '^secret/.*version:2' ; then
        vault secrets enable -version=2 kv || \
            die "Can't enable secrets kv!"
    fi
}

function enable_approle () {
    if ! vault auth list |grep -q approle ; then
        vault auth enable approle || \
            die "Can't enable approle!"
    fi
}

function write_policy() {
    vault policy write test-app test/vault/vault-policy.hcl || \
        die "Can't apply test policy!"
}

function create_approle_app () {
    vault write auth/approle/role/test-app \
        policies=test-app \
        secret_id_num_uses=0 \
        secret_id_ttl=0 \
        token_max_ttl=3m \
        token_num_uses=0 \
        token_ttl=3m || \
            die "Can't create test-app approle!"
}

function get_role_id () {
    vault read auth/approle/role/test-app/role-id \
        |grep role_id \
        |awk '{print $2}'
}

function get_secret_id () {
    vault write -f auth/approle/role/test-app/secret-id \
        |grep secret_id \
        |grep -v accessor \
        |awk '{print $2}'
}

function register_app () {
    local role_id=$1
    local secret_id=$2

    vault write auth/approle/login \
        role_id=${role_id} \
        secret_id=${secret_id} || \
            die "Can't register role-id and secret-id for login!"
}

function clean_up_dot_env () {
    cat .env |grep -v -i vault > .env
}

#
# Main
#

wait_for_vault
enable_secrets_kv
enable_approle
write_policy
create_approle_app

ROLE_ID="$(get_role_id)"
[ -z $ROLE_ID ] && \
    die "Can't obtain app-role role-id for test-app!"

SECRET_ID="$(get_secret_id)"
[ -z $SECRET_ID ] && \
    die "Can't obtain app-role secret-id for test-app!"

register_app "${ROLE_ID}" "${SECRET_ID}"

clean_up_dot_env

echo "# Writing '.env' file with Vault variables..."
cat <<EOS >> .env && cat .env
VAULT_ADDR="${VAULT_ADDR}"
VAULT_TOKEN="${VAULT_TOKEN}"
GALAXY_VAULT_ROLE_ID="${ROLE_ID}"
GALAXY_VAULT_SECRET_ID="${SECRET_ID}"
EOS
