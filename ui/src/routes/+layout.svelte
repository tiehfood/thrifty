<script lang="ts">
    import "../app.css";
    import debounce from 'debounce';
    import type { Flow, PageButton } from '$lib/types';
    import { newFlowHandlerStore } from "$lib/stores";
    import { sharedState } from './layout.svelte.js';
    import Fox from '../icons/fox.svg?component';
    import {
        Button,
        Footer,
        FooterCopyright,
        Input,
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

    let { children } = $props();

    let buttons: PageButton[] = $state([editButton])
    let clickOutsideModal = $state(false);
    let hiddenFileInputRef: HTMLInputElement;

    let newFlow: Flow = $state(getEmptyFlow());
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
        clickOutsideModal = true;
    }

    function clickClose() {
        buttons = [editButton]
        sharedState.isEditMode = false;
    }

    function clickModalClose() {
        clickOutsideModal = false
        newFlow = getEmptyFlow()
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
        if (newFlow.amount === 0) {
            inputRef.value = "";
        }
    }

    async function handleSubmit(event: Event): Promise<void> {
        event.preventDefault();
        const errors = validateForm(newFlow);
        if (errors.length > 0) {
            alert(errors.join("\n"));
            return;
        }
        if (newFlowHandler) {
            await newFlowHandler(newFlow);
        }
        clickOutsideModal = false;
        newFlow = getEmptyFlow();
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
                newFlow.icon = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svgContent.toString())}`;
            }, 500);
            reader.readAsText(uploadedFile);
        }
    }
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
<Modal title="Add a new entry" bind:open={clickOutsideModal} outsideclose>
    <form onsubmit={handleSubmit}>
        <Label class="space-y-2 mb-6">
            <span>Name</span>
            <Input type="text" placeholder="Enter name" bind:value={newFlow.name}/>
        </Label>
        <Label class="space-y-2 mb-6">
            <span>Description</span>
            <Input type="text" placeholder="Enter description (optional)" bind:value={newFlow.description}/>
        </Label>
        <div class="grid grid-cols-2 gap-8">
            <Label class="space-y-2 mb-6">
                <span>Amount</span>
                <NumberInput on:focus={handleFocus} id="numberInput" step="0.01" bind:value={newFlow.amount}/>
            </Label>
            <div class="flex justify-center">
                <input type="file" accept=".svg" onchange={handleFileUpload} class="hidden" bind:this={hiddenFileInputRef} />
                <button type="button" onclick={openFileDialog} class="relative flex {newFlow.icon === undefined ? 'items-center' : ''} justify-center rounded p-1 ring-2 ring-gray-300 {newFlow.icon === undefined ? 'bg-gray-100' : 'bg-white'} text-gray-600 aspect-square h-24 w-24 text-center m-1 flex flex-shrink-0">
                    {#if (newFlow.icon === undefined)}
                        <span>Icon (optional)</span>
                    {:else }
                        <img src={newFlow.icon} alt="Icon" />
                    {/if}
                </button>
            </div>

        </div>
        <div class="flex p-0 pt-4 space-x-3">
            <Button onclick={modalAddButton.clickHandle}
                    color={modalAddButton.color}
                    type="submit">{modalAddButton.name}
            </Button>
            <Button onclick={modalCloseButton.clickHandle}
                    color={modalCloseButton.color}>
                {modalCloseButton.name}
            </Button>
        </div>
    </form>
</Modal>

<Footer>
    <hr class="my-6 border-gray-200 sm:mx-auto lg:my-8" />
    <FooterCopyright href="/" by="tiehfood" copyrightMessage="| All Rights Reserved"/>
</Footer>
