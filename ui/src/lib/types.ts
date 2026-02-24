import type { MouseEventHandler } from "svelte/elements";

export interface Flow {
    id:          string | undefined;
    name:        string;
    description: string;
    amount:      number;
    icon?:       string;
}

export interface User {
    id:   string;
    name: string;
}

export interface Settings {
    multiUserEnabled: boolean;
    currentUser: User;
}

export interface PageButton {
    name:         string;
    clickHandle?: MouseEventHandler<HTMLButtonElement>;
    color?:       "alternative" | "red" | "yellow" | "green" | "purple" | "blue" | "light" | "dark" | "primary";
    hidden?:      boolean;
}
