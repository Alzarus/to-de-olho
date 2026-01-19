# =============================================================================
# Setup Infraestrutura GCP - To De Olho
# Execute este script uma vez para configurar a infraestrutura
# =============================================================================

# Configurar projeto
gcloud config set project to-de-olho

Write-Host "Habilitando APIs necessarias..." -ForegroundColor Cyan
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable sqladmin.googleapis.com
gcloud services enable secretmanager.googleapis.com
gcloud services enable artifactregistry.googleapis.com

Write-Host "Criando Artifact Registry..." -ForegroundColor Cyan
gcloud artifacts repositories create todeolho-images `
    --repository-format=docker `
    --location=southamerica-east1 `
    --description="Docker images for To De Olho"

Write-Host "Criando instancia Cloud SQL PostgreSQL..." -ForegroundColor Cyan
Write-Host "ATENCAO: Isso pode levar alguns minutos..." -ForegroundColor Yellow
gcloud sql instances create todeolho-db `
    --database-version=POSTGRES_15 `
    --tier=db-f1-micro `
    --region=southamerica-east1 `
    --storage-size=10GB `
    --storage-auto-increase

# Solicitar senha do usuario
$dbPassword = Read-Host -Prompt "Digite a senha para o usuario postgres" -AsSecureString
$dbPasswordPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [Runtime.InteropServices.Marshal]::SecureStringToBSTR($dbPassword)
)

Write-Host "Configurando senha do PostgreSQL..." -ForegroundColor Cyan
gcloud sql users set-password postgres `
    --instance=todeolho-db `
    --password="$dbPasswordPlain"

Write-Host "Criando banco de dados..." -ForegroundColor Cyan
gcloud sql databases create todeolho --instance=todeolho-db

# Solicitar chave da API de Transparencia
$apiKey = Read-Host -Prompt "Digite a chave da API do Portal da Transparencia"

Write-Host "Criando secret no Secret Manager..." -ForegroundColor Cyan
echo $apiKey | gcloud secrets create transparencia-api-key --data-file=-

Write-Host "Criando Service Account..." -ForegroundColor Cyan
gcloud iam service-accounts create todeolho-backend `
    --display-name="To De Olho Backend"

Write-Host "Configurando permissoes..." -ForegroundColor Cyan
gcloud projects add-iam-policy-binding to-de-olho `
    --member="serviceAccount:todeolho-backend@to-de-olho.iam.gserviceaccount.com" `
    --role="roles/cloudsql.client"

gcloud secrets add-iam-policy-binding transparencia-api-key `
    --member="serviceAccount:todeolho-backend@to-de-olho.iam.gserviceaccount.com" `
    --role="roles/secretmanager.secretAccessor"

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "Infraestrutura criada com sucesso!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "Proximos passos:" -ForegroundColor Yellow
Write-Host "1. Configure os secrets no GitHub:" -ForegroundColor White
Write-Host "   - GCP_SA_KEY: JSON da service account" -ForegroundColor Gray
Write-Host "   - DB_PASSWORD: $dbPasswordPlain" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Para gerar a chave da service account, execute:" -ForegroundColor White
Write-Host "   gcloud iam service-accounts keys create sa-key.json \" -ForegroundColor Gray
Write-Host "     --iam-account=todeolho-backend@to-de-olho.iam.gserviceaccount.com" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Faca push para a branch master para iniciar o deploy" -ForegroundColor White
