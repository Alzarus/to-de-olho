// Package acesso conta visitantes unicos por dia sem guardar IP nem cookie.
//
// Cada visita vira um hash de (sal do dia, IP, user agent). O sal e aleatorio,
// vale um dia e e apagado quando o dia acaba: depois disso ninguem, nem quem
// tem acesso ao banco, consegue ligar um hash a um IP. E o mesmo desenho do
// Plausible. O total exibido na home e a soma de visitantes unicos por dia.
package acesso

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Visita e um visitante unico num dia.
type Visita struct {
	Dia  time.Time `gorm:"type:date;primaryKey"`
	Hash []byte    `gorm:"type:bytea;primaryKey"`
}

func (Visita) TableName() string { return "acessos_visitas" }

// Sal e o segredo do dia. So existe enquanto o dia nao acabou.
type Sal struct {
	Dia time.Time `gorm:"type:date;primaryKey"`
	Sal []byte    `gorm:"type:bytea;not null"`
}

func (Sal) TableName() string { return "acessos_sal" }

// Resumo alimenta o card da home.
type Resumo struct {
	Total int64      // visitantes unicos somados dia a dia
	Desde *time.Time // primeiro dia contado; nil sem nenhuma visita
}

// brasilia define a virada do dia. O Brasil nao tem horario de verao desde 2019.
var brasilia = time.FixedZone("BRT", -3*60*60)

// Repository grava e le as visitas.
type Repository struct {
	db    *gorm.DB
	agora func() time.Time
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db, agora: time.Now}
}

// Registrar conta a visita uma vez por dia para o mesmo IP e user agent.
func (r *Repository) Registrar(ip, userAgent string) error {
	dia := diaDe(r.agora())
	sal, err := r.salDoDia(dia)
	if err != nil {
		return err
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&Visita{Dia: dia, Hash: hashVisitante(sal, ip, userAgent)}).Error
}

// Resumo devolve o total de visitantes unicos por dia e o primeiro dia contado.
func (r *Repository) Resumo() (Resumo, error) {
	var linha struct {
		Total int64
		Desde *time.Time
	}
	err := r.db.Model(&Visita{}).Select("COUNT(*) AS total, MIN(dia) AS desde").Scan(&linha).Error
	return Resumo{Total: linha.Total, Desde: linha.Desde}, err
}

// salDoDia devolve o sal do dia, criando se preciso, e apaga os sais velhos.
func (r *Repository) salDoDia(dia time.Time) ([]byte, error) {
	novo := make([]byte, 32)
	if _, err := rand.Read(novo); err != nil {
		return nil, fmt.Errorf("gerar sal: %w", err)
	}
	var sal Sal
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dia < ?", dia).Delete(&Sal{}).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&Sal{Dia: dia, Sal: novo}).Error; err != nil {
			return err
		}
		return tx.Where("dia = ?", dia).First(&sal).Error
	})
	if err != nil {
		return nil, fmt.Errorf("sal do dia: %w", err)
	}
	if len(sal.Sal) == 0 {
		return nil, errors.New("sal do dia vazio")
	}
	return sal.Sal, nil
}

func diaDe(t time.Time) time.Time {
	l := t.In(brasilia)
	return time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)
}

func hashVisitante(sal []byte, ip, userAgent string) []byte {
	h := sha256.New()
	h.Write(sal)
	h.Write([]byte{0})
	h.Write([]byte(ip))
	h.Write([]byte{0})
	h.Write([]byte(userAgent))
	return h.Sum(nil)
}

// marcasDeRobo cobre os robos que executam JavaScript (Googlebot, Bingbot) e
// os navegadores automatizados. Robo sem JavaScript nem chega a chamar a rota.
var marcasDeRobo = []string{"bot", "crawl", "spider", "slurp", "headless", "lighthouse", "preview", "python", "curl", "wget", "go-http-client"}

// EhRobo diz se o user agent e de robo ou esta vazio.
func EhRobo(userAgent string) bool {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		return true
	}
	for _, m := range marcasDeRobo {
		if strings.Contains(ua, m) {
			return true
		}
	}
	return false
}
