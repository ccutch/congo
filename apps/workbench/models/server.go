package models

// type Server struct {
// 	congo.Model
// 	Name      string
// 	Domain    string
// 	Region    string
// 	IpAddress string
// 	Error     string
// }

// func AllServers(db *congo.Database) ([]*Server, error) {
// 	var servers []*Server
// 	return servers, db.Query(`

// 		SELECT id, name, domain_name, ip_address, error, region, created_at, updated_at
// 		FROM servers
// 		ORDER BY created_at DESC

// 	`).All(func(scan congo.Scanner) error {
// 		s := Server{Model: congo.Model{DB: db}}
// 		servers = append(servers, &s)
// 		return scan(&s.ID, &s.Name, &s.Domain, &s.IpAddress, &s.Error, &s.Region, &s.CreatedAt, &s.UpdatedAt)
// 	})
// }

// func NewServer(db *congo.Database, name, region, domain string) (*Server, error) {
// 	s := Server{Model: db.NewModel(uuid.NewString())}
// 	return &s, db.Query(`

// 		INSERT INTO servers (id, name, region, domain_name)
// 		VALUES (?, ?, ?, ?)
// 		RETURNING name, domain_name, ip_address, error, region, created_at, updated_at

// 	`, s.ID, name, region, domain).Scan(&s.Name, &s.Domain, &s.IpAddress, &s.Error, &s.Region, &s.CreatedAt, &s.UpdatedAt)
// }

// func GetServer(db *congo.Database, id string) (*Server, error) {
// 	s := Server{Model: congo.Model{DB: db}}
// 	return &s, db.Query(`

// 		SELECT id, name, domain_name, ip_address, error, region, created_at, updated_at
// 		FROM servers
// 		WHERE id = ?

// 	`, id).Scan(&s.ID, &s.Name, &s.Domain, &s.IpAddress, &s.Error, &s.Region, &s.CreatedAt, &s.UpdatedAt)
// }

// func (s *Server) Save() error {
// 	return s.DB.Query(`

// 		UPDATE servers
// 		SET name = ?,
// 			  region = ?,
// 			  domain_name = ?,
// 				ip_address = ?,
// 				error = ?,
// 				updated_at = CURRENT_TIMESTAMP
// 		WHERE id = ?
// 		RETURNING updated_at

// 	`, s.Name, s.Region, s.Domain, s.IpAddress, s.Error, s.ID).Scan(&s.UpdatedAt)
// }

// func (s *Server) Delete() error {
// 	return s.DB.Query(`

// 		DELETE FROM servers
// 		WHERE id = ?

// 	`, s.ID).Exec()
// }
