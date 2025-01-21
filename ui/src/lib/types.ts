export interface Flow {
    id:          string | undefined;
    name:        string;
    description: string;
    amount:      number;
    icon?:       string;
}

export interface PageButton {
    name:         string;
    clickHandle?: Function;
    color?:       string;
    hidden?:      boolean;
}
