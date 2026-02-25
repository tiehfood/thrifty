import type { Settings } from '$lib/types';
import { sharedState } from '$lib/sharedState.svelte';

export const settings = $state<Settings>({
    multiUserEnabled: false,
    currentUser: {
        id: "",
        name: ""
    },
    numberFormat: 'eu-decimal',
});

export async function loadSettings(): Promise<void> {
    try {
        const res = await fetch(`${sharedState.apiBase}/api/settings`);
        if (!res.ok) return;
        const s: Settings = await res.json();
        settings.multiUserEnabled = s.multiUserEnabled;
        settings.currentUser = s.currentUser;
        settings.numberFormat = s.numberFormat;
        sharedState.multiUserEnabled = s.multiUserEnabled;
    } catch (e) {
        console.error("Error loading settings:", e);
    }
}

export async function patchSettings(
    patch: {
        multiUserEnabled?: boolean;
        currentUserId?: string;
        numberFormat?: string;
    }
): Promise<void> {
    try {
        const wasMultiUser = settings.multiUserEnabled;
        const res = await fetch(`${sharedState.apiBase}/api/settings`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(patch),
        });
        if (!res.ok) return;
        await loadSettings();
        if (patch.currentUserId !== undefined || (patch.multiUserEnabled === false && wasMultiUser)) {
            sharedState.reloadTrigger++;
        }
    } catch (e) {
        console.error("Error saving settings:", e);
    }
}
