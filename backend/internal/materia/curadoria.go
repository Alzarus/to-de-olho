package materia

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// apelidosCurados e a curadoria do projeto: materias votadas no mandato, de
// nome notorio, sem apelido oficial no Senado (conferido em 23/09/2026).
//
// Regra: cada nome aparece literalmente, ligado ao numero da materia, na
// pagina oficial de fonte_url (Agencia Senado). Nome que nao foi confirmado
// numa pagina oficial nao entra. O apelido oficial, quando o Senado passar a
// atribuir um, tem prioridade sobre este.
var apelidosCurados = []ApelidoCurado{
	{
		CodigoMateria: 158930, Identificacao: "PEC 45/2019",
		Apelido:    "Reforma tributária",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2023/11/08/senado-aprova-reforma-tributaria-no-primeiro-turno-no-plenario",
		Observacao: "Agência Senado: \"a reforma tributária (PEC 45/2019)\". Promulgada como EC 132/2023.",
	},
	{
		CodigoMateria: 164914, Identificacao: "PLP 68/2024",
		Apelido:    "Regulamentação da reforma tributária",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2024/12/12/reforma-tributaria-aprovado-o-texto-base-da-regulamentacao-sobre-consumo",
		Observacao: "Agência Senado: \"Regulamentação da reforma tributária sobre consumo é aprovada no Senado\" (relator do PLP 68/2024).",
	},
	{
		CodigoMateria: 166095, Identificacao: "PLP 108/2024",
		Apelido:    "Regulamentação da reforma tributária (segunda parte)",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2025/09/30/modificada-regulamentacao-da-reforma-tributaria-volta-a-camara",
		Observacao: "Agência Senado: \"PLP 108/2024 que regulamenta a segunda parte da reforma tributária\"; cria o Comitê Gestor do IBS.",
	},
	{
		CodigoMateria: 148030, Identificacao: "PEC 8/2021",
		Apelido:    "PEC que limita decisões monocráticas",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2023/11/22/senado-aprova-pec-que-limita-decisoes-individuais-em-tribunais",
		Observacao: "Agência Senado: \"a PEC 8/2021, que limita decisões monocráticas (individuais)\".",
	},
	{
		CodigoMateria: 157888, Identificacao: "PL 2903/2023",
		Apelido:    "Marco temporal para terras indígenas",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2023/09/27/aprovado-no-senado-marco-temporal-para-terras-indigenas-segue-para-sancao",
		Observacao: "Agência Senado: \"Aprovado no Senado, marco temporal para terras indígenas segue para sanção\" (PL 2.903/2023).",
	},
	{
		CodigoMateria: 160148, Identificacao: "PEC 48/2023",
		Apelido:    "PEC do marco temporal",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2025/12/09/aprovada-em-dois-turnos-pec-do-marco-temporal-vai-a-camara",
		Observacao: "Agência Senado: \"Aprovada em dois turnos, PEC do marco temporal vai à Câmara\" (PEC 48/2023).",
	},
	{
		CodigoMateria: 160011, Identificacao: "PEC 45/2023",
		Apelido:    "PEC sobre drogas",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2024/04/16/senado-aprova-pec-sobre-drogas-que-segue-para-a-camara",
		Observacao: "Agência Senado: \"a PEC sobre drogas. A PEC 45/2023 insere no art. 5º...\".",
	},
	{
		CodigoMateria: 154451, Identificacao: "PL 2253/2022",
		Apelido:    "Restrição à saída temporária de presos",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2024/02/20/senado-aprova-restricao-as-saidinhas-de-presos-texto-volta-para-a-camara",
		Observacao: "Agência Senado: \"PL 2.253/2022 que restringe o benefício da saída temporária para presos\" (o \"saidão\").",
	},
	{
		CodigoMateria: 148785, Identificacao: "PL 2159/2021",
		Apelido:    "Lei Geral do Licenciamento Ambiental",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2025/05/21/senado-aprova-projeto-da-lei-do-licenciamento-ambiental",
		Observacao: "Agência Senado: \"o projeto que cria a Lei Geral do Licenciamento Ambiental (LGLA). O PL 2.159/2021...\".",
	},
	{
		CodigoMateria: 166801, Identificacao: "PL 4932/2024",
		Apelido:    "Restrição ao uso de celulares nas escolas",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2024/12/18/senado-aprova-restricao-do-uso-de-celulares-por-estudantes-em-escolas",
		Observacao: "Agência Senado: \"Senado aprova restrição do uso de celulares por estudantes em escolas\" (PL 4.932/2024). Virou a Lei 15.100/2025.",
	},
	{
		CodigoMateria: 164599, Identificacao: "PLP 121/2024",
		Apelido:    "Propag",
		FonteURL:   "https://www12.senado.leg.br/noticias/materias/2024/08/14/senado-aprova-renegociacao-de-dividas-dos-estados-com-a-uniao",
		Observacao: "Agência Senado: \"Programa de Pleno Pagamento de Dívidas dos Estados (Propag)\" (PLP 121/2024). O substitutivo da Câmara já tem o apelido oficial \"Propag\".",
	},
}

// ApelidosCurados devolve uma copia da lista curada (para testes e relatorios).
func ApelidosCurados() []ApelidoCurado {
	return append([]ApelidoCurado(nil), apelidosCurados...)
}

// SemearApelidosCurados grava a curadoria (upsert pelo codigo_materia) e
// remove o que saiu da lista. Idempotente: roda no startup.
func SemearApelidosCurados(db *gorm.DB) error {
	return semear(db, apelidosCurados)
}

func semear(db *gorm.DB, lista []ApelidoCurado) error {
	return db.Transaction(func(tx *gorm.DB) error {
		codigos := make([]int, 0, len(lista))
		for _, a := range lista {
			codigos = append(codigos, a.CodigoMateria)
		}
		del := tx.Where("1 = 1")
		if len(codigos) > 0 {
			del = tx.Where("codigo_materia NOT IN ?", codigos)
		}
		if err := del.Delete(&ApelidoCurado{}).Error; err != nil {
			return fmt.Errorf("remover apelidos fora da curadoria: %w", err)
		}
		if len(lista) == 0 {
			return nil
		}
		copia := append([]ApelidoCurado(nil), lista...)
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "codigo_materia"}},
			DoUpdates: clause.AssignmentColumns([]string{"identificacao", "apelido", "fonte_url", "observacao", "updated_at"}),
		}).Create(&copia).Error
		if err != nil {
			return fmt.Errorf("gravar apelidos curados: %w", err)
		}
		slog.Debug("apelidos curados gravados", "total", len(copia))
		return nil
	})
}
