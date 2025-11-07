package main

type Database interface {
	Connect() string
	Query(sql string) string
}

type MySQL struct {
	Host string
}

func (m MySQL) Connect() string {
	return "Connected to MySQL at " + m.Host
}

func (m MySQL) Query(sql string) string {
	return "MySQL: " + sql
}

type PostgreSQL struct {
	Host string
}

func (p PostgreSQL) Connect() string {
	return "Connected to PostgreSQL at " + p.Host
}

func (p PostgreSQL) Query(sql string) string {
	return "PostgreSQL: " + sql
}

func NewDatabase(dbType, host string) Database {
	switch dbType {
	case "mysql":
		return MySQL{Host: host}
	case "postgres":
		return PostgreSQL{Host: host}
	default:
		return nil
	}
}

func main() {}
