<script lang="ts">
    import "../app.css";
    import debounce from 'debounce';
    import type { Flow, PageButton } from '$lib/types';
    import { newFlowHandlerStore, editFlowHandlerStore } from "$lib/stores";
    import { sharedState } from './layout.svelte.js';
    import Fox from '../icons/fox.svg?component';
    import {
        Button,
        Footer,
        FooterCopyright,
        Input,
        Img,
        Label,
        Modal,
        Navbar,
        NavBrand,
        NumberInput
    } from "flowbite-svelte";

    const editButton: PageButton        = { name: "Edit", clickHandle: clickEdit, color: "alternative" }
    const newButton: PageButton         = { name: "New Entry", clickHandle: clickNew }
    const dummyButton: PageButton       = { name: "", hidden: true }
    const closeButton: PageButton       = { name: "Close", clickHandle: clickClose, color: "alternative" }
    const modalCloseButton: PageButton  = { name: "Close", clickHandle: clickModalClose, color: "alternative" }
    const modalAddButton: PageButton    = { name: "Add" }
    const modalEditButton: PageButton    = { name: "Edit" }

    let { children } = $props();

    let buttons: PageButton[] = $state([editButton])
    let clickOutsideModal = $state(false);
    let hiddenFileInputRef: HTMLInputElement;

    let currentFlow: Flow = $state(getEmptyFlow());    
    let newFlowHandler: (flow: Flow) => void;
    newFlowHandlerStore.subscribe((handler: (flow: Flow) => void) => newFlowHandler = handler);

    function getEmptyFlow(): Flow {
        return {
            id: undefined,
            name: "",
            description: "",
            amount: 0.0,
            icon: undefined
        }
    }

    function clickEdit() {
        buttons = [dummyButton, newButton, closeButton];
        sharedState.isEditMode = true;
    }

    function clickNew() {
        currentFlow = getEmptyFlow();
        clickOutsideModal = true;
    }

    function clickClose() {
        buttons = [editButton];
        sharedState.isEditMode = false;
    }

    function clickModalClose() {
        clickOutsideModal = false;
    }

    function validateForm(flow: Flow): string[] {
        const errors: string[] = [];
        if (!flow.name.trim()) {
            errors.push("Name is required.");
        }
        if (flow.amount === 0) {
            errors.push("Amount should not be zero.");
        }
        return errors;
    }

    function handleFocus() {
        let inputRef = document.getElementById("numberInput") as HTMLInputElement;
        if (currentFlow.amount === 0) {
            inputRef.value = "";
        }
    }

    async function handleSubmit(event: Event): Promise<void> {
        event.preventDefault();
        const errors = validateForm(currentFlow);
        if (errors.length > 0) {
            alert(errors.join("\n"));
            return;
        }
        if (newFlowHandler) {
            await newFlowHandler(currentFlow);
        }
        clickOutsideModal = false;
        currentFlow = getEmptyFlow();
    }

    function openFileDialog() {
        hiddenFileInputRef.click();
    }

    function handleFileUpload(event: Event) {
        const uploadedFile = (event.target as HTMLInputElement).files?.[0];
        if (uploadedFile) {
            const reader = new FileReader();
            reader.onload = debounce((e) => {
                const svgContent = e.target.result;
                currentFlow.icon = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svgContent.toString())}`;
            }, 500);
            reader.readAsText(uploadedFile);
        }
    }

    async function editFlowHandler(flow: Flow) {
        currentFlow = flow;
        clickOutsideModal = true;
    }

    editFlowHandlerStore.set(editFlowHandler);
</script>

<style>
    :global(body) {
        background-color: #DDD;
    }
</style>

<svelte:head>
    <title>Thrifty</title>
    <link rel="apple-touch-icon" href="./icons/apple-touch-icon.png" />
    <link rel="manifest" href="./manifest.webmanifest" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
</svelte:head>

{#snippet addButton(button: PageButton, type?: "submit" | "reset" | "button" | null | undefined)}
    <Button onclick={button.clickHandle}
            color={button.color}
            {type}>{button.name}
    </Button>
{/snippet}

<Navbar class="bg-gray-700 text-gray-200">
    <NavBrand href="/">
        <Fox class="h-12 me-3 p-1 sm:h-16" alt="Thrifty" />
        <span class="self-center whitespace-nowrap text-2xl sm:text-3xl text-primary-800 font-medium">Thrifty</span>
    </NavBrand>
    <div class="ml-auto flex">
        {#each buttons as button}
            <Button onclick={button.clickHandle}
                    class="me-3 px-4 sm:px-5 py-2 sm:py-2.5 {button.hidden ? 'hidden' : ''}"
                    color={button.color}>
                {button.name}
            </Button>
        {/each}
    </div>
</Navbar>

{@render children()}
<Modal title={currentFlow.id ? "Edit entry" : "Add new entry"} bind:open={clickOutsideModal} outsideclose backdropClass="fixed inset-0 z-40 bg-gray-900/50">
    <form onsubmit={handleSubmit}>
        <Label class="space-y-2 mb-6">
            <span>Name</span>
            <Input type="text" placeholder="Enter name" bind:value={currentFlow.name}/>
        </Label>
        <Label class="space-y-2 mb-6">
            <span>Description</span>
            <Input type="text" placeholder="Enter description (optional)" bind:value={currentFlow.description}/>
        </Label>
        <div class="grid grid-cols-2 gap-8">
            <Label class="space-y-2 mb-6">
                <span>Amount</span>
                <NumberInput onfocus={handleFocus} id="numberInput" step="0.01" bind:value={currentFlow.amount}/>
            </Label>
            <div class="flex justify-center">
                <input type="file" accept=".svg" onchange={handleFileUpload} class="hidden" bind:this={hiddenFileInputRef} />
                <button type="button" onclick={openFileDialog} class="relative flex {currentFlow.icon === undefined ? 'items-center' : ''} justify-center rounded p-1 ring-2 ring-gray-300 {currentFlow.icon === undefined ? 'bg-gray-100' : 'bg-white'} aspect-square h-24 w-24 m-1 flex-shrink-0 cursor-pointer">
                    {#if (currentFlow.icon === undefined)}
                        <span class="text-center text-gray-600">Icon (optional)</span>
                    {:else }
                        <Img size="w-full h-full" class="object-contain" src={currentFlow.icon} alt="Icon" />
                    {/if}
                </button>
            </div>

        </div>
        <div class="flex p-0 pt-4 space-x-3">
            {#if currentFlow.id}
                {@render addButton(modalEditButton, "submit")}
            {:else}
                {@render addButton(modalAddButton, "submit")}
            {/if}
            {@render addButton(modalCloseButton)}
        </div>
    </form>
</Modal>

<Footer>
    <hr class="my-6 border-gray-200 sm:mx-auto lg:my-8" />
    <FooterCopyright href="https://github.com/tiehfood/thrifty" by="tiehfood" copyrightMessage="| All Rights Reserved"/>
</Footer>
