import type { Icon } from '$lib/types';
import { sharedState } from '$lib/sharedState.svelte';

export const icons = $state<Icon[]>([]);

export async function loadIcons(): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/icons`);
        if (!res.ok) return;
        const fresh: Icon[] = await res.json();
        icons.splice(0, icons.length, ...fresh);
    } catch (e) {
        console.error("Error loading icons:", e);
    }
}

export async function addIcons(dataUris: string[]): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/icons`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(dataUris.map(data => ({ data }))),
        });
        if (!res.ok) return;
        await loadIcons();
    } catch (e) {
        console.error("Error adding icons:", e);
    }
}

export async function removeIcon(id: string): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/icons/${id}`, {
            method: "DELETE",
        });
        if (!res.ok) return;
        await loadIcons();
    } catch (e) {
        console.error("Error deleting icon:", e);
    }
}
