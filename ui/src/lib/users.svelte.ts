import type { User } from '$lib/types';
import { loadSettings } from '$lib/settings.svelte';
import { sharedState } from '$lib/sharedState.svelte';

export const users = $state<User[]>([]);

export async function loadUsers(): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/users`);
        if (!res.ok) return;
        const fresh: User[] = await res.json();
        users.splice(0, users.length, ...fresh);
    } catch (e) {
        console.error("Error loading users:", e);
    }
}

export async function createUser(name: string): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/users`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name }),
        });
        if (!res.ok) return;
        await loadSettings();
        await loadUsers();
        sharedState.reloadTrigger++;
    } catch (e) {
        console.error("Error creating user:", e);
    }
}

export async function updateUser(id: string, name: string): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/users/${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name }),
        });
        if (!res.ok) return;
        await loadSettings();
    } catch (e) {
        console.error("Error updating user:", e);
    }
}

export async function removeUser(id: string): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/users/${id}`, {
            method: "DELETE"
        });
        if (!res.ok) return;
        await loadSettings();
        await loadUsers();
        sharedState.reloadTrigger++;
    } catch (e) {
        console.error("Error deleting user:", e);
    }
}
