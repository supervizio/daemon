# GPG Key Setup for Package Signing

This guide explains how to set up GPG keys for signing supervizio packages.

## Generate GPG Key

```bash
# Generate a new GPG key (RSA 4096-bit, no expiration for CI)
gpg --full-generate-key

# Select:
# - (1) RSA and RSA
# - 4096 bits
# - 0 = key does not expire
# - Real name: supervizio
# - Email: noreply@superviz.io
# - Comment: Package Signing Key
```

## Export Keys

```bash
# Get key ID
KEY_ID=$(gpg --list-secret-keys --keyid-format LONG | grep sec | head -1 | awk '{print $2}' | cut -d'/' -f2)
echo "Key ID: $KEY_ID"

# Export signing key (for GitHub Secrets)
gpg --armor --export-secret-keys "$KEY_ID" > supervizio-signing.asc

# Export public key (for distribution)
gpg --armor --export "$KEY_ID" > supervizio-public.asc

# Get fingerprint
gpg --fingerprint "$KEY_ID"
```

## Configure GitHub Secrets

Add the following secrets to your GitHub repository:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add new repository secrets:

| Secret Name | Value |
|-------------|-------|
| `GPG_SIGNING_KEY` | Contents of `supervizio-signing.asc` |
| `GPG_PASSPHRASE` | Your GPG passphrase (if set) |

## Enable GitHub Pages

1. Go to **Settings** → **Pages**
2. Source: **GitHub Actions**
3. The `deploy-repo.yml` workflow will automatically deploy

## Verify Setup

After a release is published:

1. Check the workflow run in **Actions** → **Deploy Repository**
2. Visit `https://<owner>.github.io/<repo>/` to see the repository page
3. Test installation:

```bash
# Test GPG key import
curl -fsSL https://supervizio.github.io/daemon/gpg.key | gpg --show-keys

# Test APT repository
curl -fsSL https://supervizio.github.io/daemon/apt/dists/stable/Release
```

## Key Rotation

To rotate the GPG key:

1. Generate a new key (see above)
2. Update GitHub Secrets with new key
3. Trigger a new release or manual workflow run
4. Update documentation with new fingerprint

## Security Best Practices

- Store the signing key backup in a secure location (password manager, HSM)
- Use a strong passphrase for the GPG key
- Consider using a dedicated signing key (not your personal key)
- Enable 2FA on your GitHub account
- Review Actions audit logs periodically

## Troubleshooting

### "No secret key" error

```bash
# Check if key is imported
gpg --list-secret-keys

# Re-import if needed
gpg --import supervizio-signing.asc
```

### Package signature verification fails

```bash
# Check key fingerprint
gpg --fingerprint <KEY_ID>

# Verify the public key matches
curl -fsSL https://supervizio.github.io/daemon/gpg.key | gpg --show-keys
```

### RPM signing fails

```bash
# Check ~/.rpmmacros
cat ~/.rpmmacros

# Should contain:
# %_gpg_name <KEY_ID>
```
