import type { MouseEventHandler } from "svelte/elements";

export interface Flow {
    id:          string | undefined;
    name:        string;
    description: string;
    amount:      number;
    icon?:       string;
}

export interface Icon {
    id:     string;
    data:   string;
    isUsed: boolean;
}

export interface User {
    id:   string;
    name: string;
}

export interface Settings {
    multiUserEnabled: boolean;
    currentUser: User;
    numberFormat: string;
}

export interface NumberFormatOption {
    value:   string;
    name:    string;
    locales: string;
    format:  Intl.NumberFormatOptions;
}

export interface PageButton {
    name:         string;
    clickHandle?: MouseEventHandler<HTMLButtonElement>;
    color?:       "alternative" | "red" | "yellow" | "green" | "purple" | "blue" | "light" | "dark" | "primary";
    hidden?:      boolean;
}
