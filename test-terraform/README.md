# Testing Local Terraform Provider

This directory contains test Terraform configurations for testing the local SigNoz provider.

## Prerequisites

1. **Go** (version 1.21 or higher)
2. **Terraform** (version 1.0 or higher)
3. **SigNoz** instance running (default: http://localhost:3301)
4. **SigNoz API Key** - Get this from your SigNoz instance

## Getting Your SigNoz API Key

1. Log in to your SigNoz instance (e.g., http://localhost:3301)
2. Go to **Settings** → **API Keys**
3. Create a new API key or use an existing one
4. Copy the API key

## Quick Start - All Commands

Follow these commands in order:

### Step 1: Update Configuration

Edit `main.tf` and replace `<YOUR-SIGNOZ-API-KEY>` with your actual API key:

```bash
# Open the file
nano main.tf  # or use your preferred editor

# Find and replace this line:
# access_token = "<YOUR-SIGNOZ-API-KEY>"
# with your actual key:
# access_token = "your_actual_api_key_here"
```

### Step 2: Build and Install the Local Provider

Run the setup script from this directory:

```bash
cd /Users/pranavgupta/GolandProjects/terraform-provider-signoz/test-terraform
./setup-local-provider.sh
```

This script will:
- Build the provider from your local code
- Install it to `~/.terraform.d/plugins/`
- Make it available for Terraform to use

### Step 3: Initialize Terraform

```bash
terraform init
```

This will initialize Terraform and recognize your local provider.

### Step 4: Review the Terraform Plan

```bash
terraform plan
```

This shows what resources will be created without actually creating them.

### Step 5: Apply the Configuration

```bash
terraform apply
```

Type `yes` when prompted to create the alerts.

### Step 6: View Created Resources

```bash
terraform show
```

Or check the outputs:

```bash
terraform output
```

### Step 7: Verify in SigNoz UI

1. Open your SigNoz instance: http://localhost:3301
2. Go to **Alerts** section
3. You should see the 3 created alerts:
   - High Memory Usage
   - Critical CPU Usage
   - High Error Rate in Logs

## Making Changes

After modifying the provider code:

### Option A: Quick Rebuild and Test

```bash
# 1. Rebuild and reinstall the provider
./setup-local-provider.sh

# 2. Remove Terraform lock file to use the new provider
rm .terraform.lock.hcl

# 3. Reinitialize
terraform init -upgrade

# 4. Test your changes
terraform plan
```

### Option B: Clean Start

```bash
# 1. Destroy existing resources
terraform destroy

# 2. Clean Terraform state
rm -rf .terraform terraform.tfstate*

# 3. Rebuild provider
./setup-local-provider.sh

# 4. Start fresh
terraform init
terraform apply
```

## Manual Build and Install (Alternative Method)

If the setup script doesn't work, follow these manual steps:

```bash
# 1. Navigate to project root
cd /Users/pranavgupta/GolandProjects/terraform-provider-signoz

# 2. Build the provider
go build -o terraform-provider-signoz_v0.0.1

# 3. Determine your OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Convert arch (x86_64 -> amd64, arm64 stays arm64)
if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi

# 4. Create plugin directory
mkdir -p ~/.terraform.d/plugins/signoz.local/local/signoz/0.0.1/${OS}_${ARCH}

# 5. Copy and make executable
cp terraform-provider-signoz_v0.0.1 ~/.terraform.d/plugins/signoz.local/local/signoz/0.0.1/${OS}_${ARCH}/
chmod +x ~/.terraform.d/plugins/signoz.local/local/signoz/0.0.1/${OS}_${ARCH}/terraform-provider-signoz_v0.0.1

# 6. Cleanup
rm terraform-provider-signoz_v0.0.1

# 7. Go back to test directory
cd test-terraform

# 8. Initialize and run
terraform init
terraform plan
terraform apply
```

## Cleanup - Destroy Resources

To remove all created alerts:

```bash
terraform destroy
```

Type `yes` when prompted.

## Useful Terraform Commands

```bash
# Format your Terraform files
terraform fmt

# Validate configuration
terraform validate

# Show current state
terraform show

# List all resources
terraform state list

# Show specific resource
terraform state show signoz_alert.high_memory

# Import existing alert (if needed)
terraform import signoz_alert.high_memory <alert-id>

# See detailed logs
TF_LOG=DEBUG terraform apply
```

## Troubleshooting

### Provider Not Found

If you get "provider not found" error:

```bash
# Remove lock file and reinit
rm .terraform.lock.hcl
terraform init -upgrade
```

### Authentication Error

Check:
1. SigNoz is running: `curl http://localhost:3301`
2. API key is correct in `main.tf`
3. API key has necessary permissions

### Build Errors

```bash
# Clean and rebuild
cd /Users/pranavgupta/GolandProjects/terraform-provider-signoz
go clean
go mod tidy
go build
```

### Provider Version Conflict

```bash
# Clean everything and start fresh
cd test-terraform
rm -rf .terraform .terraform.lock.hcl terraform.tfstate*
./setup-local-provider.sh
terraform init
```

## Example Alert Condition Structures

The alerts in `main.tf` include examples of:
- **Metric-based alerts** with builder queries
- **Threshold rules** with targets
- **Label and channel configurations**
- **Different severity levels**

To create custom alerts:
1. Create an alert in SigNoz UI
2. Use `terraform import` to import it
3. Run `terraform show` to see its structure
4. Use that structure as a template

## File Structure

```
test-terraform/
├── main.tf                    # Terraform configuration with alerts
├── setup-local-provider.sh    # Script to build and install provider
└── README.md                  # This file
```

## Next Steps

1. Modify the alerts in `main.tf` to match your metrics
2. Add more alerts as needed
3. Test different alert types (LOGS_BASED_ALERT, TRACES_BASED_ALERT, etc.)
4. Configure alert channels and notifications

## Additional Resources

- [Terraform Provider Documentation](../docs/)
- [SigNoz Documentation](https://signoz.io/docs/)
- [Provider Source Code](../signoz/)