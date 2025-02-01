package models

type GeneralProductivity struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	Ano                int    `json:"ano"`
	ParlamentarAutor   string `json:"parlamentar_autor"`
	TotalProjetosDeLei int    `json:"total_projetos_de_lei"`
	TotalRequerimentos int    `json:"total_requerimentos"`
	TotalIndicacoes    int    `json:"total_indicacoes"`
	Outros             int    `json:"outros"`
	Total              int    `json:"total"`
	Tipo               string `json:"tipo"`
}
