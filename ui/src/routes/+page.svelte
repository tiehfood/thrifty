<script lang="ts">
    import type { Flow, Group } from "$lib/types";
    import { Badge, Card, Indicator, Modal, Button, Img } from "flowbite-svelte";
    import NumberFlow, { continuous } from "@number-flow/svelte";

    import { onMount } from "svelte";
    import { newFlowHandlerStore, editFlowHandlerStore, NUMBER_FORMATS } from "$lib/stores.js";
    import { sharedState } from "./layout.svelte.js";
    import { settings } from "$lib/settings.svelte";

    const currentFormat = $derived(NUMBER_FORMATS.find(f => f.value === settings.numberFormat) ?? NUMBER_FORMATS[0]);

    let flows: Flow[] = $state([]);
    let groups: Group[] = $state([]);
    let total: number = $state(0);
    let shouldDelete: boolean = $state(false);

    let groupModalOpen = $state(false);
    let selectedGroup = $state<Group | null>(null);
    let groupFlows = $state<Flow[]>([]);
    let shouldDeleteGroupFlow = $state(false);

    let editFlowHandler: (flow: Flow) => void;
    editFlowHandlerStore.subscribe((handler: (flow: Flow) => void) => editFlowHandler = handler);

    $effect(() => {
        if (sharedState.reloadTrigger > 0) refreshAll();
    });

    onMount(async () => {
        await refreshAll();
    });

    async function refreshAll() {
        await Promise.all([getFlows(), getGroups()]);
        if (groupModalOpen && selectedGroup?.id) {
            await fetchGroupFlows(selectedGroup.id);
        }
    }

    async function getFlows() {
        try {
            let response = await fetch(`${sharedState.apiBase}/api/flows`);
            if (!response.ok) throw new Error(response.statusText);
            flows = await response.json();
            setTotal();
        } catch (error) {
            console.error("Error fetching flows:", error);
        }
    }

    async function getGroups() {
        try {
            let response = await fetch(`${sharedState.apiBase}/api/groups`);
            if (!response.ok) throw new Error(response.statusText);
            const data: Group[] = await response.json();
            groups.splice(0, groups.length, ...data.filter(g => g.entryCount > 0));
            setTotal();
        } catch (error) {
            console.error("Error fetching groups:", error);
        }
    }

    async function fetchGroupFlows(groupId: string) {
        try {
            const res = await fetch(`${sharedState.apiBase}/api/groups/${groupId}/flows`);
            if (!res.ok) throw new Error(res.statusText);
            const data: Flow[] = await res.json();
            groupFlows.splice(0, groupFlows.length, ...data);
        } catch (error) {
            console.error("Error fetching group flows:", error);
        }
    }

    async function openGroupModal(group: Group) {
        selectedGroup = group;
        await fetchGroupFlows(group.id as string);
        groupModalOpen = true;
    }

    async function uploadFlow(flow: Flow) {
        try {
            const response = await fetch(`${sharedState.apiBase}/api/flows`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(flow),
            });
            if (!response.ok) throw new Error(response.statusText);
        } catch (error) {
            console.error("Error uploading flow:", error);
        }
    }

    async function updateFlow(flow: Flow) {
        try {
            const response = await fetch(`${sharedState.apiBase}/api/flows/${flow.id}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(flow),
            });
            if (!response.ok) throw new Error(response.statusText);
        } catch (error) {
            console.error("Error updating flow:", error);
        }
    }

    async function deleteFlow(id: string) {
        try {
            const response = await fetch(`${sharedState.apiBase}/api/flows/${id}`, {
                method: "DELETE",
            });
            if (!response.ok) throw new Error(response.statusText);
        } catch (error) {
            console.error("Error deleting flow:", error);
        }
    }

    async function newFlowHandler(flow: Flow) {
        if (flow.id) {
            await updateFlow(flow);
        } else {
            await uploadFlow(flow);
        }
        await refreshAll();
    }

    async function deleteFlowHandler(id: string) {
        await deleteFlow(id);
        await refreshAll();
        shouldDelete = false;
    }

    async function deleteGroupFlowHandler(id: string) {
        await deleteFlow(id);
        await refreshAll();
        shouldDeleteGroupFlow = false;
    }

    function setTotal() {
        const flowSum = flows.reduce((acc, f) => acc + f.amount, 0);
        const groupSum = groups.reduce((acc, g) => acc + g.amount, 0);
        total = Math.round((flowSum + groupSum + Number.EPSILON) * 100) / 100;
    }

    type Item = { kind: 'flow'; data: Flow } | { kind: 'group'; data: Group };

    function getIncome(flows: Flow[], groups: Group[]): Item[] {
        return [
            ...flows.filter(f => f.amount > 0).map(f => ({ kind: 'flow' as const, data: f })),
            ...groups.filter(g => g.amount > 0).map(g => ({ kind: 'group' as const, data: g })),
        ].sort((a, b) => b.data.amount - a.data.amount);
    }

    function getExpenses(flows: Flow[], groups: Group[]): Item[] {
        return [
            ...flows.filter(f => f.amount <= 0).map(f => ({ kind: 'flow' as const, data: f })),
            ...groups.filter(g => g.amount <= 0).map(g => ({ kind: 'group' as const, data: g })),
        ].sort((a, b) => a.data.amount - b.data.amount);
    }

    function shouldUseColumns(flows: Flow[], groups: Group[]): string {
        return (getIncome(flows, groups).length === 0 || getExpenses(flows, groups).length === 0) ? "" : "md:grid-cols-2 gap-4";
    }

    newFlowHandlerStore.set(newFlowHandler);
</script>

{#snippet flowItem(flow: Flow)}
    <div class="p-1">
        <Card size="sm" class="mx-auto relative p-3">
            <div class="flex items-center space-x-4 rtl:space-x-reverse px-2">
                <img class="justify-center rounded-none w-10 h-10 flex-shrink-0 bg-white" src={flow.icon} alt="Icon" />
                <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-gray-900 truncate">{flow.name}</p>
                    <p class="text-xs text-gray-500 truncate">{flow.description}</p>
                </div>
                <div class="inline-flex items-center text-base font-semibold text-gray-900">
                    <NumberFlow locales={currentFormat.locales} animated={false} format={currentFormat.format} value={flow.amount} />
                </div>
            </div>
            {#if sharedState.isEditMode}
                <button type="button" onclick={() => { if (!shouldDelete) editFlowHandler({ ...flow }) }} class="absolute inset-0 flex items-center justify-center rounded-lg border -m-[1px] !border-gray-700 {shouldDelete ? 'bg-red-800/85' : 'bg-green-800/85'} opacity-0 hover:opacity-100 cursor-pointer">
                    <span class="text-lg font-bold text-gray-100 truncate">{shouldDelete ? 'Delete' : 'Edit'}</span>
                    <Indicator class="bg-red-600 hover:bg-red-700" border onmouseenter={() => shouldDelete = true} onmouseleave={() => shouldDelete = false} onclick={() => deleteFlowHandler(flow.id as string)} size="xl" placement="top-right">
                        <span class="text-white text-xs font-bold">—</span>
                    </Indicator>
                </button>
            {/if}
        </Card>
    </div>
{/snippet}

{#snippet groupItem(group: Group)}
    <div class="p-1">
        <Card size="sm" class="mx-auto relative p-3 cursor-pointer" onclick={() => openGroupModal(group)}>
            <div class="flex items-center space-x-4 rtl:space-x-reverse px-2">
                <img class="justify-center rounded-none w-10 h-10 flex-shrink-0 bg-white" src={group.icon} alt="Icon" />
                <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-gray-900 truncate">{group.name}</p>
                    <p class="text-xs text-gray-500 truncate">{group.description}</p>
                </div>
                <div class="flex flex-col items-end gap-1">
                    <div class="inline-flex items-center text-base font-semibold text-gray-900">
                        <NumberFlow locales={currentFormat.locales} animated={false} format={currentFormat.format} value={group.amount} />
                    </div>
                    <Badge color="pink" class="text-xs">Group</Badge>
                </div>
            </div>
        </Card>
    </div>
{/snippet}

<div class="p-8">
    <div class="p-1">
        <Card size="sm" class="mx-auto p-6">
            <div class="flex items-center space-x-4 rtl:space-x-reverse">
                <div class="flex-1 min-w-0">
                    <p class="text-lg font-bold text-gray-900 truncate">Total</p>
                </div>
                <div class="inline-flex items-center text-base font-black text-gray-900">
                    <NumberFlow locales={currentFormat.locales} plugins={[continuous]} format={currentFormat.format} value={total} />
                </div>
            </div>
        </Card>
    </div>
    {#if flows.length > 0 || groups.length > 0}
        <hr class="h-0.5 my-6 bg-gray-700 border-0">
    {/if}
    <div class="grid grid-cols-1 {shouldUseColumns(flows, groups)}">
        <div>
            {#each getIncome(flows, groups) as item}
                {#if item.kind === 'flow'}
                    {@render flowItem(item.data)}
                {:else}
                    {@render groupItem(item.data)}
                {/if}
            {/each}
        </div>
        <div>
            {#each getExpenses(flows, groups) as item}
                {#if item.kind === 'flow'}
                    {@render flowItem(item.data)}
                {:else}
                    {@render groupItem(item.data)}
                {/if}
            {/each}
        </div>
    </div>
</div>

<Modal title={selectedGroup?.name ?? 'Group'} bind:open={groupModalOpen} outsideclose>
    <div class="{groupFlows.length > 5 ? 'max-h-[320px] overflow-y-auto' : ''}">
        {#each groupFlows as flow}
            <div class="p-1">
                <Card size="sm" class="mx-auto relative p-3">
                    <div class="flex items-center space-x-4 rtl:space-x-reverse px-2">
                        <img class="justify-center rounded-none w-10 h-10 flex-shrink-0 bg-white" src={flow.icon} alt="Icon" />
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium text-gray-900 truncate">{flow.name}</p>
                            <p class="text-xs text-gray-500 truncate">{flow.description}</p>
                        </div>
                        <div class="inline-flex items-center text-base font-semibold text-gray-900">
                            <NumberFlow locales={currentFormat.locales} animated={false} format={currentFormat.format} value={flow.amount} />
                        </div>
                    </div>
                    {#if sharedState.isEditMode}
                        <button type="button" onclick={() => { if (!shouldDeleteGroupFlow) editFlowHandler({ ...flow }) }} class="absolute inset-0 flex items-center justify-center rounded-lg border -m-[1px] !border-gray-700 {shouldDeleteGroupFlow ? 'bg-red-800/85' : 'bg-green-800/85'} opacity-0 hover:opacity-100 cursor-pointer">
                            <span class="text-lg font-bold text-gray-100 truncate">{shouldDeleteGroupFlow ? 'Delete' : 'Edit'}</span>
                            <Indicator class="bg-red-600 hover:bg-red-700" border onmouseenter={() => shouldDeleteGroupFlow = true} onmouseleave={() => shouldDeleteGroupFlow = false} onclick={() => deleteGroupFlowHandler(flow.id as string)} size="xl" placement="top-right">
                                <span class="text-white text-xs font-bold">—</span>
                            </Indicator>
                        </button>
                    {/if}
                </Card>
            </div>
        {/each}
    </div>
    <div class="flex p-0 pt-4">
        <Button onclick={() => groupModalOpen = false} color="alternative">Close</Button>
    </div>
</Modal>
