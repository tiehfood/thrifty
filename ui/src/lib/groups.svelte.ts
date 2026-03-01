import type { Group } from "$lib/types";
import { sharedState } from "$lib/sharedState.svelte";

export const groups = $state<Group[]>([]);

export async function loadGroups(): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/groups`);
        if (!res.ok) throw new Error(res.statusText);
        const data: Group[] = await res.json();
        groups.splice(0, groups.length, ...data);
    } catch (error) {
        console.error("Error loading groups:", error);
    }
}

export async function createGroup(group: Omit<Group, 'id' | 'amount'>): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/groups`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(group),
        });
        if (!res.ok) throw new Error(res.statusText);
        await loadGroups();
    } catch (error) {
        console.error("Error creating group:", error);
    }
}

export async function updateGroup(id: string, group: Omit<Group, 'id' | 'amount'>): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/groups/${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(group),
        });
        if (!res.ok) throw new Error(res.statusText);
        await loadGroups();
    } catch (error) {
        console.error("Error updating group:", error);
    }
}

export async function removeGroup(id: string): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/groups/${id}`, {
            method: "DELETE",
        });
        if (!res.ok) throw new Error(res.statusText);
        await loadGroups();
    } catch (error) {
        console.error("Error removing group:", error);
    }
}
