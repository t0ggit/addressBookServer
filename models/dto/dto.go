package dto

type Cond struct {
	Lop    string
	PgxInd string
	Field  string
	Value  any
}

// todo | ну тут либо разбираться с тем, что он написал (конды и вся эта шаблон генерация)
// todo | либо поискать какой-нибудь готовый инструмент для таких рофлов
// todo | но видимо он хочет, чтобы сами придумали решение с помощью text/template, а не готовое
// todo | ну крч походу все-таки надо разбираться в его реализации и как-то мб ее допилить
