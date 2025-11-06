# Quick Start - Run These Commands

## Step-by-Step Commands to Run Your Local Terraform Provider

### 1. Get Your SigNoz API Key First

Before running any commands, get your API key:
- Open: http://localhost:3301
- Go to: Settings → API Keys
- Copy your API key

### 2. Update the Configuration

```bash
cd /Users/pranavgupta/GolandProjects/terraform-provider-signoz/test-terraform

# Edit main.tf and replace <YOUR-SIGNOZ-API-KEY> with your actual key
nano main.tf
# or
open -e main.tf
```

### 3. Build and Install Local Provider

```bash
./setup-local-provider.sh
```

Expected output:
```
========================================
Setting up Local Terraform Provider
========================================
...
✓ Provider built successfully
✓ Plugin directory created
✓ Provider installed
✓ Installation Complete!
========================================
```

### 4. Initialize Terraform

```bash
terraform init
```

Expected output:
```
Initializing provider plugins...
- Finding signoz.local/local/signoz versions matching "0.0.1"...
- Installing signoz.local/local/signoz v0.0.1...
...
Terraform has been successfully initialized!
```

### 5. Preview Changes

```bash
terraform plan
```

This shows what will be created (3 alerts).

### 6. Create the Alerts

```bash
terraform apply
```

Type `yes` when prompted.

Expected output:
```
...
signoz_alert.high_memory: Creating...
signoz_alert.critical_cpu: Creating...
signoz_alert.high_error_rate: Creating...
...
Apply complete! Resources: 3 added, 0 changed, 0 destroyed.
```

### 7. Verify

Check the outputs:
```bash
terraform output
```

Or verify in SigNoz UI:
- Open: http://localhost:3301/alerts
- You should see 3 new alerts

### 8. Clean Up (Optional)

To remove all created alerts:
```bash
terraform destroy
```

Type `yes` when prompted.

---

## After Making Code Changes

When you modify the provider code and want to test:

```bash
# 1. Rebuild and reinstall
./setup-local-provider.sh

# 2. Update Terraform
rm .terraform.lock.hcl
terraform init -upgrade

# 3. Test
terraform plan
terraform apply
```

---

## Troubleshooting

**Error: Provider not found**
```bash
rm .terraform.lock.hcl
terraform init -upgrade
```

**Error: Authentication failed**
- Check SigNoz is running: `curl http://localhost:3301`
- Verify API key in `main.tf`

**Error: Build failed**
```bash
cd /Users/pranavgupta/GolandProjects/terraform-provider-signoz
go mod tidy
cd test-terraform
./setup-local-provider.sh
```

---

## All Commands in One Block (Copy-Paste Ready)

```bash
# Navigate to test directory
cd /Users/pranavgupta/GolandProjects/terraform-provider-signoz/test-terraform

# Edit main.tf to add your API key (do this first!)
nano main.tf

# Build and install provider
./setup-local-provider.sh

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Create alerts
terraform apply
# Type: yes

# Verify outputs
terraform output

# Check in browser
open http://localhost:3301/alerts

# Clean up (when done testing)
# terraform destroy
```