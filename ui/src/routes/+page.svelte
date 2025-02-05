<script lang="ts">
    import type { Flow } from "$lib/types";
    import { Card, Indicator } from "flowbite-svelte";
    import NumberFlow, { continuous } from '@number-flow/svelte';

    import { onMount } from "svelte";
    import { newFlowHandlerStore, editFlowHandlerStore } from "$lib/stores.js";
    import { sharedState } from "./layout.svelte.js";

    let flows: Flow[] = $state([]);
    let total: number = $state(0);
    let shouldDelete: boolean = $state(false);

    let editFlowHandler: (flow: Flow) => void;
    editFlowHandlerStore.subscribe((handler: (flow: Flow) => void) => editFlowHandler = handler);

    let currentProtocol: string;
    let currentHostname: string;
    let currentPort: string;

    onMount(async () => {
        currentProtocol = window.location.protocol.replace(":","");
        currentHostname = window.location.hostname.toLowerCase();
        currentPort = window.location.port.trim();
        await getFlows();
    })

    async function getFlows() {
        try {
            let response = await fetch(`${currentProtocol}://${currentHostname}${!!currentPort ? ":" : ""}${currentPort}/api/flows`);
            if (!response.ok) throw new Error(response.statusText);
            flows = await response.json();
            setTotal();
        } catch (error) {
            console.error("Error fetching flows:", error);
        }
    }

    async function uploadFlow(flow: Flow) {
        try {
            //console.log(JSON.stringify(flow));
            const response = await fetch(`${currentProtocol}://${currentHostname}${!!currentPort ? ":" : ""}${currentPort}/api/flows`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(flow),
            });
            if (!response.ok) throw new Error(response.statusText);
        } catch (error) {
            console.error("Error uploading flow:", error);
        }
    }

    async function updateFlow(flow: Flow) {
        try {
            const response = await fetch(`${currentProtocol}://${currentHostname}${!!currentPort ? ":" : ""}${currentPort}/api/flows/${flow.id}`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(flow),
            });
            if (!response.ok) throw new Error(response.statusText);
        } catch (error) {
            console.error("Error uploading flow:", error);
        }
    }    

    async function deleteFlow(id: string) {
        try {
            const response = await fetch(`${currentProtocol}://${currentHostname}${!!currentPort ? ":" : ""}${currentPort}/api/flows/${id}`, {
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
        await getFlows();
    }

    async function deleteFlowHandler(id: string) {
        await deleteFlow(id);
        await getFlows();
    }

    function setTotal() {
        let sum = flows.reduce((accumulator, currentFlow) => accumulator + currentFlow.amount, 0);
        total = Math.round((sum + Number.EPSILON) * 100) / 100;
    }

    function getExpenses(flows: Flow[]): Flow[] {
        return flows.filter(item => item.amount <= 0).reverse()
    }

    function getIncome(flows: Flow[]): Flow[] {
        return flows.filter(item => item.amount > 0)
    }

    function shouldUseColumns(flows: Flow[]): string {
        return (getIncome(flows).length == 0 || getExpenses(flows).length == 0) ? "" : "md:grid-cols-2 gap-4"
    }

    newFlowHandlerStore.set(newFlowHandler);
</script>

{#snippet flowCard(flows: Flow[])}
    <div>
        {#each flows as flow}
            <div class="p-1">
                <Card padding="xs" size="sm" class="mx-auto relative">
                    <div class="flex items-center space-x-4 rtl:space-x-reverse px-2">
                        <img class="justify-center rounded-none w-10 h-10 flex-shrink-0 bg-white" src={flow.icon} alt="Icon" />
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium text-gray-900 truncate">
                                {flow.name}
                            </p>
                            <p class="text-xs text-gray-500 truncate">
                                {flow.description}
                            </p>
                        </div>
                        <div class="inline-flex items-center text-base font-semibold text-gray-900">
                            <NumberFlow locales="de-DE" animated={false} format={{ style: 'currency', currency: 'EUR', trailingZeroDisplay: 'stripIfInteger' }} value={flow.amount} />
                        </div>
                    </div>
                    {#if (sharedState.isEditMode)}
                        <button type="button" onclick={() => {if (!shouldDelete) editFlowHandler({ ...flow})}} class="absolute inset-0 flex items-center justify-center rounded-lg border -m-[1px] !border-gray-700 {shouldDelete ? 'bg-red-800/85' : 'bg-green-800/85'} opacity-0 hover:opacity-100 cursor-pointer">
                            <span class="text-lg font-bold text-gray-100 truncate">{shouldDelete ? 'Delete' : 'Edit'}</span>
                            <Indicator class="bg-red-600 hover:bg-red-700" border onmouseenter={() => shouldDelete = true} onmouseleave={() => shouldDelete = false} onclick={() => deleteFlowHandler(flow.id as string)} size="xl" placement="top-right">
                                <span class="text-white text-xs font-bold">—</span>
                            </Indicator>
                        </button>
                    {/if}
                </Card>
            </div>
        {/each}
    </div>
{/snippet}

<div class="p-8">
    <div class="p-1">
        <Card padding="lg" size="sm" class="mx-auto">
            <div class="flex items-center space-x-4 rtl:space-x-reverse">
                <div class="flex-1 min-w-0">
                    <p class="text-lg font-bold text-gray-900 truncate">
                        Total
                    </p>
                </div>
                <div class="inline-flex items-center text-base font-black text-gray-900">
                    <NumberFlow locales="de-DE" plugins={[continuous]} format={{ style: 'currency', currency: 'EUR', trailingZeroDisplay: 'stripIfInteger' }} value={total} />
                </div>
            </div>
        </Card>
    </div>
    {#if (flows.length > 0)}
        <hr class="h-0.5 my-6 bg-gray-700 border-0">
    {/if}
    <div class="grid grid-cols-1 {shouldUseColumns(flows)}">
        {@render flowCard(getIncome(flows))}
        {@render flowCard(getExpenses(flows))}
    </div>
</div>
