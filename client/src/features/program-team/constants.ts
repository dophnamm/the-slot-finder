/** Under the program's key, so accepting a draft (which refetches programs) refreshes the team's hours. */
export const programTeamQueryKey = (id: string) => ['programs', id, 'team'] as const

/** Everyone a planner may add to a team. */
export const TEAM_CANDIDATES_QUERY_KEY = ['program-team', 'candidates'] as const
