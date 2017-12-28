package model

import "github.com/google/gofuzz"

func (e *NewsEntity) Fuzz(c fuzz.Continue) {
	e.Title = randNonEmptyString(c)
	e.Content = c.RandString()
	e.Continue = c.RandBool()
	//e.CreatedAt=
	if !e.Lead.Valid {
		c.Fuzz(&e.Lead)
	}
	e.Score = c.Rand.Float64()
	e.Version = c.Rand.Int63n(100)
	if !e.ViewsDistribution.Valid {
		c.Fuzz(&e.ViewsDistribution)
	}
}

func (e *CategoryEntity) Fuzz(c fuzz.Continue) {
	e.Name = randNonEmptyString(c)
	e.Content = c.RandString()
	if !e.ParentID.Valid {
		c.Fuzz(&e.ParentID)
	}
}

func randNonEmptyString(c fuzz.Continue) (res string) {
Start:
	res = c.RandString()
	if res == "" {
		goto Start
	}
	return
}
