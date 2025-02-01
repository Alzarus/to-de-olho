package models

type PropositionProductivity struct {
	ID                               uint   `json:"id" gorm:"primaryKey"`
	Ano                              int    `json:"ano"`
	ParlamentarAutor                 string `json:"parlamentar_autor"`
	Mocao                            *int   `json:"mocao"`
	ProjetoDecretoLegislativo        *int   `json:"projeto_decreto_legislativo"`
	ProjetoEmendaLOM                 *int   `json:"projeto_emenda_lom"`
	ProjetoIndicacao                 *int   `json:"projeto_indicacao"`
	ProjetoLeiComplementar           *int   `json:"projeto_lei_complementar"`
	ProjetoLei                       *int   `json:"projeto_lei"`
	ProjetoResolucao                 *int   `json:"projeto_resolucao"`
	RequerimentoAdministrativo       *int   `json:"requerimento_administrativo"`
	RequerimentoUrgenciaUrgentissima *int   `json:"requerimento_urgencia_urgentissima"`
	RequerimentoUtilidadePublica     *int   `json:"requerimento_utilidade_publica"`
	RequerimentoEspecial             *int   `json:"requerimento_especial"`
	Veto                             *int   `json:"veto"`
	Total                            int    `json:"total"`
	Tipo                             string `json:"tipo"`
}
