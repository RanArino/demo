## AWS MSK (Serverless) – Implementation Guide for Scaler Demo

This guide prepares AWS resources and runtime settings so you can run the demo containers against Amazon MSK Serverless and verify event streaming.

- Target region: set to your chosen AWS Region (examples use `ap-northeast-1`).
- Demo topics: `document.uploaded`, `document.processed`
- Consumer groups: `ms_knowledge`, `ms_document_process`

### Prerequisites
- AWS account with permissions for MSK, IAM, EC2, and Secrets Manager (optional).
- AWS CLI v2 installed and configured (`aws configure`) or SSO login.
- A VPC with at least two private subnets across distinct AZs.
- Java 11+ for Kafka CLI on a client host (only needed if you want to create topics/test from CLI).

## 1) Create an MSK Serverless cluster (Console)
1. Open Amazon MSK → Create cluster → choose "Serverless".
2. Name the cluster (e.g., `msk-serverless-knowledge`).
3. Networking:
   - Select your VPC and at least two private subnets across AZs.
   - Select a security group for broker ENIs (outbound allowed; inbound typically default/no special inbound needed for brokers).
4. Security:
   - Encryption in transit: TLS (default).
   - Authentication: enable IAM (SASL/IAM).
5. Create and wait for Status: Active.
"""
(result)
First VPC configuration
Subnets
subnet-038001d0ce50c5f28 
subnet-05e4ea72af40221e6 
subnet-0b186f0ac0a5cb551

Security groups applied
sg-0afeca9ded12c91e8 
"""


Reference values you will need after creation:
- Cluster ARN: visible on the cluster details page.
- Cluster UUID: embedded in ARNs (e.g., `.../cluster/<name>/<uuid>-s2`).
- Bootstrap brokers (SASL/IAM): in "View client information"; use these for `KAFKA_BROKERS`.

CLI to list/retrieve brokers (optional):
```bash
aws kafka list-clusters-v2 --query "ClusterInfoList[].ClusterArn" --output table
aws kafka get-bootstrap-brokers \
  --cluster-arn <your-cluster-arn> \
  --query 'BootstrapBrokerStringSaslIam' \
  --output text
```

## 2) IAM – Least-privilege data plane access for clients
MSK IAM authorization evaluates the client principal (IAM user/role) when it connects. You do not attach a role to the MSK cluster. Instead, attach an IAM policy to the client principal.

### 2.1 Create policy
Create a customer-managed policy similar to the following (replace placeholders with your Region, Account ID, cluster name, cluster UUID, topics, and consumer groups):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["kafka-cluster:Connect", "kafka-cluster:DescribeCluster"],
      "Resource": "arn:aws:kafka:<region>:<account-id>:cluster/<cluster-name>/<cluster-uuid>"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kafka-cluster:DescribeTopic",
        "kafka-cluster:CreateTopic",
        "kafka-cluster:WriteData",
        "kafka-cluster:ReadData"
      ],
      "Resource": [
        "arn:aws:kafka:<region>:<account-id>:topic/<cluster-name>/<cluster-uuid>/document.uploaded",
        "arn:aws:kafka:<region>:<account-id>:topic/<cluster-name>/<cluster-uuid>/document.processed"
      ]
    },
    {
      "Effect": "Allow",
      "Action": ["kafka-cluster:AlterGroup", "kafka-cluster:DescribeGroup"],
      "Resource": "arn:aws:kafka:<region>:<account-id>:group/<cluster-name>/<cluster-uuid>/*"
    }
  ]
}
```

Notes:
- Omit `kafka-cluster:AlterCluster` for least privilege unless you need to change cluster settings.
- For early testing, you can temporarily use wildcards for topic and group resource ARNs, then tighten later.

### 2.2 Create a client principal
Choose one:
- Local dev with static keys: Create an IAM user (programmatic access). Attach the policy above.
- EC2/ECS/EKS runtime (recommended for MSK): Create an IAM role and attach the policy. Attach the role to the compute (EC2 instance profile, ECS task role, or EKS service account via IRSA).

=== DONE ===

## 3) Networking – where to run the demo
MSK Serverless brokers are private; there are no public endpoints. Your demo containers must run in the same VPC (or a connected network).

Options:
- Recommended: Launch a small EC2 instance in the cluster VPC/private subnets. Run the demo containers there.
- Alternative: Use your platform of choice in the VPC (ECS/EKS). For a quick demo, EC2 is fastest.

Security groups:
- The EC2 instance security group should allow outbound to the broker ENIs (default outbound-all is fine). Inbound SSH from your IP for management.

## 4) Prepare application environment variables
Collect from MSK → "View client information":
- KAFKA_BROKERS: the SASL/IAM bootstrap brokers string (port 9098).
- AWS_REGION: the region of your MSK cluster.

Set in the demo root `.env` (used by `docker compose` variable interpolation):
```bash
KAFKA_BROKERS=<comma-separated SASL/IAM brokers>
AWS_REGION=<your-region>
```

If you use an IAM user instead of an EC2 role, also set on the EC2 host (or export in shell):
```bash
AWS_ACCESS_KEY_ID=<key>
AWS_SECRET_ACCESS_KEY=<secret>
AWS_SESSION_TOKEN=<token-if-temporary>
```

Optional: mirror `KAFKA_BROKERS` and `AWS_REGION` in `ms_knowledge/.env.local` and `ms_document_process/.env.local` if you run services standalone.

## 5) Topic creation (one-time) from a VPC client
To create topics using the Kafka CLI with MSK IAM auth:

### 5.1 Install Kafka CLI and MSK IAM auth lib on the EC2 client
```bash
# Java 11+, then download Kafka that matches your MSK Kafka version
KAFKA_VERSION=3.6.0
curl -L -o kafka.tgz https://downloads.apache.org/kafka/$KAFKA_VERSION/kafka_2.13-$KAFKA_VERSION.tgz
mkdir -p $HOME/kafka && tar -xzf kafka.tgz -C $HOME/kafka --strip-components=1

# Download AWS MSK IAM auth library (refer to AWS docs for the latest URL)
# Example (version may change):
mkdir -p $HOME/msk-iam && cd $HOME/msk-iam
curl -L -o aws-msk-iam-auth.jar "https://github.com/aws/aws-msk-iam-auth/releases/download/v1.1.9/aws-msk-iam-auth-1.1.9-all.jar"
```

Create `client.properties`:
```properties
security.protocol=SASL_SSL
sasl.mechanism=AWS_MSK_IAM
sasl.jaas.config=software.amazon.msk.auth.iam.IAMLoginModule required;
sasl.client.callback.handler.class=software.amazon.msk.auth.iam.IAMClientCallbackHandler
ssl.endpoint.identification.algorithm=https
```

Then create topics:
```bash
BROKERS="<SASL/IAM brokers string>" # from View client information

$HOME/kafka/bin/kafka-topics.sh --create \
  --bootstrap-server "$BROKERS" \
  --command-config client.properties \
  --topic document.uploaded \
  --partitions 3

$HOME/kafka/bin/kafka-topics.sh --create \
  --bootstrap-server "$BROKERS" \
  --command-config client.properties \
  --topic document.processed \
  --partitions 3

# Verify
$HOME/kafka/bin/kafka-topics.sh --list \
  --bootstrap-server "$BROKERS" \
  --command-config client.properties
```

## 6) Run the demo containers on EC2
From the project root on the EC2 instance:

1. Ensure `.env` contains `KAFKA_BROKERS` and `AWS_REGION`.
2. If using IAM user credentials, export them in the shell (or use `aws configure`). If using an EC2 role, no keys are needed.
3. Start the services. For a quick start, you can run the existing compose. The local Kafka container in the compose is not used when `KAFKA_BROKERS` points to MSK, but it may still start due to `depends_on`. For a cleaner run, use an override file to disable the local Kafka service.

Example override `docker-compose.msk.yml` to disable local Kafka:
```yaml
services:
  kafka:
    profiles: ["disabled"]

  ms_knowledge:
    depends_on: []

  ms_document_process:
    depends_on: ["ms_knowledge"]
```

Run with override:
```bash
docker compose -f docker-compose.yml -f docker-compose.msk.yml up -d --build
```

If you prefer to keep local Kafka starting harmlessly, simply run:
```bash
docker compose up -d --build
```
The services will use MSK because `KAFKA_BROKERS` points to the MSK SASL/IAM endpoints.

## 7) Smoke test through the application
- Trigger the workflow that publishes to `document.uploaded`.
- Verify the `ms_document_process` consumer handles the event and publishes to `document.processed`.
- Check service logs and CloudWatch metrics for MSK.

## 8) Monitoring and troubleshooting
- Enable CloudWatch metrics and CloudWatch Logs for the cluster (Console → your cluster → Monitoring/Logging).
- Common issues:
  - Authorization: Verify the IAM policy ARNs include the correct `<cluster-name>/<cluster-uuid>` and topic/group names.
  - Networking: Ensure the EC2 instance runs in the same VPC/subnets and has outbound connectivity to brokers.
  - Brokers string: Use the SASL/IAM brokers (port 9098), not plaintext 9092.
  - Credentials: If using an IAM user, ensure env vars are exported in the shell where containers run.

## 9) Optional – Centralize config in AWS Secrets Manager
- Store `KAFKA_BROKERS` and `AWS_REGION` (and other app secrets) in a secret.
- Note the secret ARN; services that support cloud secret managers can load from there.

## 10) Cleanup (optional)
- Stop containers.
- Delete topics (Kafka CLI) if desired.
- Delete MSK cluster and associated EC2 resources when done.

---

### Quick Reference – Environment
- Required: `KAFKA_BROKERS`, `AWS_REGION`
- Local/EC2 with IAM user: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, optional `AWS_SESSION_TOKEN`
- Place in root `.env` for compose; optionally mirror in `ms_knowledge/.env.local` and `ms_document_process/.env.local`.


