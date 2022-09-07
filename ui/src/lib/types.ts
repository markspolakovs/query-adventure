export interface Team {
    id: string;
    name: string;
    color: string;
    members: string[];
}

export type Scoreboard = Record<string, number>;
export type CompletedChallenges = Record<string, Record<string, Record<string, boolean>>>;
