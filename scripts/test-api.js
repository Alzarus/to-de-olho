const axios = require('axios');

const API_BASE = 'https://dadosabertos.camara.leg.br/api/v2';

async function testAPICamara() {
    console.log('🔍 Testando API da Câmara dos Deputados...\n');

    try {
        // 1. Testar lista de deputados
        console.log('1️⃣ Buscando lista de deputados...');
        const deputadosResponse = await axios.get(`${API_BASE}/deputados?ordem=ASC&ordenarPor=nome`, {
            headers: {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
                'Accept': 'application/json',
                'Accept-Language': 'pt-BR,pt;q=0.9,en;q=0.8',
                'Cache-Control': 'no-cache',
                'Pragma': 'no-cache'
            },
            timeout: 10000
        });
        
        const deputados = deputadosResponse.data.dados.slice(0, 5); // Primeiros 5
        console.log(`✅ Sucesso! Encontrados ${deputadosResponse.data.dados.length} deputados`);
        console.log('📋 Primeiros 5 deputados:');
        deputados.forEach(dep => {
            console.log(`   - ${dep.nome} (${dep.siglaPartido}/${dep.siglaUf})`);
        });

        // 2. Testar deputado específico
        const primeiroDeputado = deputados[0];
        console.log(`\n2️⃣ Buscando detalhes de: ${primeiroDeputado.nome}`);
        
        const deputadoDetalhes = await axios.get(`${API_BASE}/deputados/${primeiroDeputado.id}`, {
            headers: {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
                'Accept': 'application/json',
                'Accept-Language': 'pt-BR,pt;q=0.9,en;q=0.8'
            },
            timeout: 10000
        });
        
        console.log('✅ Detalhes do deputado obtidos com sucesso!');
        console.log(`   📧 Email: ${deputadoDetalhes.data.dados.email || 'Não informado'}`);
        console.log(`   📱 Situação: ${deputadoDetalhes.data.dados.condicaoEleitoral}`);

        // 3. Testar despesas
        console.log(`\n3️⃣ Buscando despesas de 2025 de: ${primeiroDeputado.nome}`);
        
        const despesasResponse = await axios.get(`${API_BASE}/deputados/${primeiroDeputado.id}/despesas?ano=2025&ordem=DESC`, {
            headers: {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
                'Accept': 'application/json',
                'Accept-Language': 'pt-BR,pt;q=0.9,en;q=0.8'
            },
            timeout: 10000
        });
        
        const despesas = despesasResponse.data.dados.slice(0, 3);
        console.log(`✅ Sucesso! Encontradas ${despesasResponse.data.dados.length} despesas em 2025`);
        
        if (despesas.length > 0) {
            console.log('💰 Últimas 3 despesas:');
            despesas.forEach(despesa => {
                console.log(`   - ${despesa.tipoDespesa}: R$ ${despesa.valorLiquido.toFixed(2)}`);
            });
            
            const totalGasto = despesasResponse.data.dados.reduce((acc, d) => acc + d.valorLiquido, 0);
            console.log(`💸 Total gasto em 2025: R$ ${totalGasto.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}`);
        }

        console.log('\n🎉 TESTE CONCLUÍDO COM SUCESSO!');
        console.log('✅ A API da Câmara está funcionando corretamente');
        console.log('✅ Todos os endpoints necessários estão acessíveis');
        
    } catch (error) {
        console.error('\n❌ ERRO no teste da API:');
        console.error('🔴 Detalhes:', error.message);
        
        if (error.response) {
            console.error(`🔴 Status HTTP: ${error.response.status}`);
            console.error(`🔴 Resposta: ${JSON.stringify(error.response.data, null, 2)}`);
        }
        
        process.exit(1);
    }
}

// Executar teste
testAPICamara();
