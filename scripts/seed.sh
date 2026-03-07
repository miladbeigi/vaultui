#!/bin/sh
# Populates Vault with fake test data for local development.
# All values below are fictional — safe to commit.
set -e

sleep 2

vault audit enable file file_path=stdout

# ── App secrets (typical web app config) ──────────────────
vault kv put secret/apps/myapp/config \
  db_host=fake-db.test.internal \
  db_port=5432 \
  db_name=testdb \
  db_user=testuser \
  db_password=FAKE-password-do-not-use

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
vault write auth/userpass/users/testuser password=FAKE-pass-do-not-use policies=default

vault auth enable approle
vault write auth/approle/role/test-role token_ttl=1h token_max_ttl=4h policies=default

vault auth enable -description="LDAP auth (unconfigured)" ldap

# ── Policies ──────────────────────────────────────────────
vault policy write readonly - <<'POLICY'
path "secret/data/*" {
  capabilities = ["read", "list"]
}
POLICY

vault policy write admin - <<'POLICY'
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
POLICY

vault policy write app-myapp - <<'POLICY'
path "secret/data/apps/myapp/*" {
  capabilities = ["read", "list"]
}
path "secret/metadata/apps/myapp/*" {
  capabilities = ["list"]
}
POLICY

echo "Seed data loaded successfully."
