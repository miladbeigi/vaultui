#!/bin/sh
# Populates Vault with fake test data for local development.
# All values below are fictional — safe to commit.
set -e

sleep 2

vault audit enable file file_path=stdout

# ── App secrets (typical web app config) ──────────────────
# Version 1
vault kv put secret/apps/myapp/config \
  db_host=fake-db.test.internal \
  db_port=5432 \
  db_name=testdb \
  db_user=testuser \
  db_password=FAKE-password-do-not-use

# Version 2 — changed password and added log_level
vault kv put secret/apps/myapp/config \
  db_host=fake-db.test.internal \
  db_port=5432 \
  db_name=testdb \
  db_user=testuser \
  db_password=FAKE-new-password-v2 \
  log_level=info

# Version 3 — changed log_level, removed db_name
vault kv put secret/apps/myapp/config \
  db_host=fake-db.test.internal \
  db_port=5432 \
  db_user=testuser \
  db_password=FAKE-new-password-v2 \
  log_level=debug \
  api_endpoint=https://api.test.example.com/v2

vault kv put secret/apps/myapp/database \
  connection_url="postgresql://fake-db.test.internal:5432/testdb" \
  username=test-admin \
  password=FAKE-hunter2-not-real

vault kv put secret/apps/myapp/api-keys \
  stripe_key=sk_test_FAKE_000000000000000000000000 \
  sendgrid_key=SG.FAKE_TEST_KEY_000000000000

vault kv put secret/apps/billing/config \
  api_url=https://billing.test.example.com \
  timeout=30s \
  webhook_secret=whsec_FAKE_test_000000000000

# ── Infrastructure secrets ────────────────────────────────

# TLS: multi-line PEM cert + private key (dummy, not real)
CERT=$(printf '-----BEGIN CERTIFICATE-----\nFAKE-TEST-CERT-DO-NOT-USE-IN-PRODUCTION\nMIIFazCCA1OgAwIBAgIUEJzGFMHhKLkjSbGBbCOIZmQDM3owDQYJKoZIhvcNAQEL\nBQAwRTELMAkGA1UEBhMCVVMxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM\nGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yNTAxMDEwMDAwMDBaFw0yNjAx\nMDEwMDAwMDBaMEUxCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw\n-----END CERTIFICATE-----')
PKEY=$(printf '-----BEGIN RSA PRIVATE KEY-----\nFAKE-TEST-KEY-DO-NOT-USE-IN-PRODUCTION\nMIIEowIBAAKCAQEA2a2rwplBQLfBCW1OZMM1RqRFHGFLMND/TjKDMp1NFQT9XGBx\nXm/K3FYPbLdGKyNBCMzjQbNOIwJce2fCMNHVMk+YHaGFMI3bMGdWPGT5MBM0XzW\nwJ4c1M5y3H8E7LqB4AkFPQa3igNprWHDflAeB3Y2PWMZR8rOaYEMQm1zKHPXLF0\nhs0tXmAjr6YMPBbtGEQaFyTVppFNahJBNMKBOCfS6IYGwqfGVR7bBjLNOlF0jVaM\n-----END RSA PRIVATE KEY-----')
vault kv put secret/infra/tls/wildcard cert="$CERT" key="$PKEY"

# AWS: fictional credentials (match AWS example format)
vault kv put secret/infra/aws \
  access_key=AKIAEXAMPLE000TEST00 \
  secret_key=FAKE/TestSecretKey+DO+NOT+USE+IN+PRODUCTION0

# SSH: multi-line deploy key (dummy, not real)
SSH_PRIV=$(printf '-----BEGIN OPENSSH PRIVATE KEY-----\nFAKE-TEST-SSH-KEY-DO-NOT-USE\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACBHKJHasdkjhASDKJHASKDJHASKDJHASDKJHASDKJHASD==AAAAFQAAAA\nRkVwbG95QHByb2QBAgMEBQ==\n-----END OPENSSH PRIVATE KEY-----')
vault kv put secret/infra/ssh/deploy-key \
  public_key="ssh-rsa AAAAB3NzaFAKETESTKEY000000000000 test-deploy@fake-host" \
  private_key="$SSH_PRIV"

# ── Auth methods ───────────────────────────────────────────
vault auth enable userpass
vault write auth/userpass/users/testuser password=testpass policies=base-read,app-secrets

vault auth enable approle
vault write auth/approle/role/test-role token_ttl=1h token_max_ttl=4h policies=base-read,infra-secrets
ROLE_ID=$(vault read -field=role_id auth/approle/role/test-role/role-id)
vault write -f auth/approle/role/test-role/secret-id
echo "AppRole role_id: $ROLE_ID"

vault auth enable -description="LDAP auth (unconfigured)" ldap

# ── Policies ──────────────────────────────────────────────

# Shared base: lets any authenticated user see engines, auth, policies, health
vault policy write base-read - <<'POLICY'
path "sys/mounts" {
  capabilities = ["read"]
}
path "sys/auth" {
  capabilities = ["read"]
}
path "sys/policies/acl" {
  capabilities = ["read", "list"]
}
path "sys/policies/acl/*" {
  capabilities = ["read"]
}
path "sys/health" {
  capabilities = ["read"]
}
path "sys/audit" {
  capabilities = ["read"]
}
path "secret/metadata" {
  capabilities = ["list"]
}
path "secret/metadata/*" {
  capabilities = ["list"]
}
POLICY

vault policy write admin - <<'POLICY'
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
POLICY

# userpass: can only read app secrets (apps/myapp/*, apps/billing/*)
vault policy write app-secrets - <<'POLICY'
path "secret/data/apps/*" {
  capabilities = ["read"]
}
path "secret/metadata/apps/*" {
  capabilities = ["list"]
}
POLICY

# approle: can only read infra secrets (infra/tls/*, infra/aws, infra/ssh/*)
vault policy write infra-secrets - <<'POLICY'
path "secret/data/infra/*" {
  capabilities = ["read"]
}
path "secret/metadata/infra/*" {
  capabilities = ["list"]
}
POLICY

# ── Identity entities and groups ─────────────────────────
vault write identity/entity name="test-user-entity" policies="base-read,app-secrets"
vault write identity/entity name="admin-entity" policies="admin"
vault write identity/group name="dev-team" policies="base-read,app-secrets" type="internal"
vault write identity/group name="ops-team" policies="admin" type="internal"

# ── PKI engine ───────────────────────────────────────────
vault secrets enable pki
vault secrets tune -max-lease-ttl=87600h pki/
vault write pki/root/generate/internal \
  common_name="Test Root CA" \
  ttl=87600h
vault write pki/roles/test-role \
  allowed_domains="test.example.com" \
  allow_subdomains=true \
  max_ttl=72h
vault write pki/issue/test-role \
  common_name="app1.test.example.com" \
  ttl=24h

# ── Transit engine ───────────────────────────────────────
vault secrets enable transit
vault write -f transit/keys/my-app-key
vault write -f transit/keys/payment-key type=aes256-gcm96

echo "Seed data loaded successfully."
