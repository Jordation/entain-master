package db

const (
	sportsList    = "list"
	sportsGet     = "get"
	sportsOrderby = "ob"
)

func getSportQueries() map[string]string {
	return map[string]string{
		sportsList: `
			SELECT 
				id, 
				event_id,
				category, 
				team_1, 
				team_2, 
				visible, 
				advertised_start_time 
			FROM sports
		`,

		sportsGet: `
			SELECT 
				id, 
				event_id,
				category, 
				team_1, 
				team_2, 
				visible, 
				advertised_start_time 
			FROM sports WHERE id = %v
		`,
		sportsOrderby: `
			ORDER BY %v %v
		`,
	}
}
