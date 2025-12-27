export interface MenuItem {
    label: string;
    action: string;
    payload: string;
}

export interface MenuDTO {
    id: number;
    slug: string;
    title: string;
    items: MenuItem[];
}

export interface DynamicTable {
    table_name: string;
    display_name: string;
}
