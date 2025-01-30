# to-de-olho
Repositório contendo o projeto Tô De Olho - Tcc de Pedro Batista de Almeida Filho

Créditos à Câmara Municipal de Salvador pela disponibilidade dos dados de transparência - https://www.cms.ba.gov.br

nest g resource data/...

docker-compose -f docker-compose.prod.yml up -d --build
docker-compose -f docker-compose.prod.yml down --remove-orphans
docker system prune -a --volumes


npm run process-json -- --workers=4 >> /var/log/json-processor.log 2>&1
docker-compose logs -f

http://45.4.247.157/

gerar /dist (./api):
npx tsc

atualizar .sh para linux (com git bash)
sed -i 's/\r$//' ./crawlers/run-crawlers.sh

pegar todos os logs dos crawlers
tail -f /app/crawlers/logs/contract.log /app/crawlers/logs/councilor.log /app/crawlers/logs/frequency.log /app/crawlers/logs/general-productivity.log /app/crawlers/logs/proposition.log /app/crawlers/logs/proposition-productivity.log /app/crawlers/logs/travel-expenses.log
