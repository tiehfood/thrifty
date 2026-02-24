<script lang="ts">
    import "../app.css";
    import debounce from "debounce";
    import type { Flow, User, PageButton } from "$lib/types";
    import { newFlowHandlerStore, editFlowHandlerStore } from "$lib/stores";
    import { sharedState } from "$lib/sharedState.svelte";
    import { settings, loadSettings, patchSettings } from "$lib/settings.svelte";
    import { users, loadUsers, createUser, updateUser, removeUser } from "$lib/users.svelte";
    import Fox from "../icons/fox.svg?component";
    import Gear from "../icons/gear.svg?component";
    import Account from "../icons/account.svg?component";
    import {
        Button,
        Card,
        Footer,
        FooterCopyright,
        Indicator,
        Input,
        Img,
        Label,
        Modal,
        Navbar,
        NavBrand,
        Toggle,
    } from "flowbite-svelte";
    import { onMount } from "svelte";

    const editButton: PageButton        = { name: "Edit", clickHandle: clickEdit, color: "light" }
    const newButton: PageButton         = { name: "New Entry", clickHandle: clickNew }
    const dummyButton: PageButton       = { name: "", hidden: true }
    const closeButton: PageButton       = { name: "Close", clickHandle: clickClose, color: "light" }
    const modalCloseButton: PageButton  = { name: "Close", clickHandle: clickModalClose, color: "alternative" }
    const modalAddButton: PageButton    = { name: "Add" }
    const modalEditButton: PageButton   = { name: "Edit" }

    let { children } = $props();

    let buttons: PageButton[] = $state([editButton]);

    let flowModalOpen = $state(false);
    let hiddenFileInputRef: HTMLInputElement;
    let currentFlow: Flow = $state(getEmptyFlow());
    let newFlowHandler: (flow: Flow) => void;
    newFlowHandlerStore.subscribe((handler: (flow: Flow) => void) => newFlowHandler = handler);

    let settingsOpen = $state(false);

    let multiUserEnabledLocal = $state(false);
    $effect(() => { multiUserEnabledLocal = settings.multiUserEnabled; });

    let renameOpen = $state(false);
    let initialUserName = $state("");

    let usersOpen = $state(false);
    let shouldDeleteUser = $state(false);

    let newUserOpen = $state(false);
    let newUserName = $state("");

    onMount(async () => {
        await loadSettings();
    });

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
        flowModalOpen = true;
    }

    function clickClose() {
        buttons = [editButton];
        sharedState.isEditMode = false;
    }

    function clickModalClose() {
        flowModalOpen = false;
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
        flowModalOpen = false;
        currentFlow = getEmptyFlow();
    }

    function openFileDialog() {
        hiddenFileInputRef.click();
    }

    function handleFileUpload(event: Event) {
        const uploadedFile = (event.target as HTMLInputElement).files?.[0];
        if (uploadedFile) {
            const reader = new FileReader();
            reader.onload = debounce((event: ProgressEvent) => {
                const svgContent = (event.target as FileReader).result ?? '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><path d="m0,0v1h1V0"/></svg>';
                currentFlow.icon = `data:image/svg+xml;base64,${btoa(svgContent.toString())}`;
            }, 500);
            reader.readAsText(uploadedFile);
        }
    }

    async function editFlowHandler(flow: Flow) {
        currentFlow = flow;
        flowModalOpen = true;
    }

    editFlowHandlerStore.set(editFlowHandler);

    function handleMultiUserToggle() {
        if (multiUserEnabledLocal && !sharedState.multiUserEnabled) {
            initialUserName = settings.currentUser.name === "thrifty" ? "" : settings.currentUser.name;
            renameOpen = true;
        } else if (!multiUserEnabledLocal && sharedState.multiUserEnabled) {
            if (!confirm("This will delete all users and their data except for the current user. Continue?")) {
                multiUserEnabledLocal = true;
                return;
            }
            patchSettings({ multiUserEnabled: false });
        } else {
            patchSettings({ multiUserEnabled: multiUserEnabledLocal });
        }
    }

    async function handleRename() {
        const name = initialUserName.trim();
        if (!name) {
            alert("Please enter a name.");
            return;
        }
        await updateUser(settings.currentUser.id, name);
        await patchSettings({ multiUserEnabled: true });
        renameOpen = false;
        settingsOpen = false;
    }

    function handleRenameCancelled() {
        if (!settings.multiUserEnabled) {
            multiUserEnabledLocal = false;
        }
        renameOpen = false;
    }

    async function openUsersModal() {
        await loadUsers();
        usersOpen = true;
    }

    async function handleSelectUser(user: User) {
        if (shouldDeleteUser) return;
        if (user.id === settings.currentUser.id) return;
        await patchSettings({ currentUserId: user.id });
        usersOpen = false;
    }

    async function handleDeleteUser(id: string) {
        await removeUser(id);
        if (users.length <= 1) {
            usersOpen = false;
        }
        shouldDeleteUser = false;
    }

    async function handleAddUser() {
        const name = newUserName.trim();
        if (!name) {
            alert("Please enter a name.");
            return;
        }
        await createUser(name);
        newUserName = "";
        newUserOpen = false;
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

{#snippet addButton(button: PageButton, type?: "submit" | "reset" | "button" | null | undefined)}
    <Button onclick={button.clickHandle} color={button.color} {type}>{button.name}</Button>
{/snippet}

<Navbar class="bg-gray-700 text-gray-200">
    <NavBrand href="/">
        <Fox class="h-12 me-3 p-1 sm:h-16" />
        <span class="self-center whitespace-nowrap text-2xl sm:text-3xl text-primary-800 font-medium">Thrifty</span>
    </NavBrand>
    <div class="ml-auto flex items-center">
        {#each buttons as button}
            <Button onclick={button.clickHandle}
                    class="me-3 px-4 sm:px-5 py-2 sm:py-2.5 {button.hidden ? 'hidden' : ''}"
                    color={button.color}>
                {button.name}
            </Button>
        {/each}
        <Button onclick={() => settingsOpen = true} color="light" class="me-3 px-3 py-2 sm:py-2.5">
            <Gear class="w-5 h-5" />
        </Button>
        {#if sharedState.multiUserEnabled}
            <Button onclick={openUsersModal} color="light" class="me-3 px-3 py-2 sm:py-2.5">
                <Account class="w-5 h-5" />
            </Button>
        {/if}
    </div>
</Navbar>

{@render children()}

<Modal title={currentFlow.id ? "Edit entry" : "Add new entry"} bind:open={flowModalOpen} outsideclose>
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
                <Input type="number" onfocus={handleFocus} id="numberInput" step="0.01" bind:value={currentFlow.amount}/>
            </Label>
            <div class="flex justify-center">
                <input type="file" accept=".svg" onchange={handleFileUpload} class="hidden" bind:this={hiddenFileInputRef} />
                <button type="button" onclick={openFileDialog} class="relative flex {currentFlow.icon === undefined ? 'items-center' : ''} justify-center rounded p-1 ring-2 ring-gray-300 {currentFlow.icon === undefined ? 'bg-gray-100' : 'bg-white'} aspect-square h-24 w-24 m-1 flex-shrink-0 cursor-pointer">
                    {#if currentFlow.icon === undefined}
                        <span class="text-center text-gray-600">Icon (optional)</span>
                    {:else}
                        <Img class="object-contain w-full h-full" src={currentFlow.icon} alt="Icon" />
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
            {@render addButton(modalCloseButton, "button")}
        </div>
    </form>
</Modal>

<Modal title="Settings" bind:open={settingsOpen} outsideclose>
    <Label class="space-y-2 mb-6">
        <Toggle bind:checked={multiUserEnabledLocal} onchange={handleMultiUserToggle}>
            Multi user support
        </Toggle>
    </Label>
    <div class="flex p-0 pt-4">
        <Button onclick={() => settingsOpen = false} color="alternative">Close</Button>
    </div>
</Modal>

<Modal title="Set your name" bind:open={renameOpen} onclose={handleRenameCancelled}>
    <Label class="space-y-2 mb-6">
        <span>Your name</span>
        <Input type="text" placeholder="Enter your name" bind:value={initialUserName} />
    </Label>
    <div class="flex p-0 pt-4 space-x-3">
        <Button onclick={handleRename}>Save</Button>
        <Button onclick={handleRenameCancelled} color="alternative">Cancel</Button>
    </div>
</Modal>

<Modal title="Users" bind:open={usersOpen} outsideclose>
    <div>
        {#each users as user}
            <div class="p-1">
                <Card size="sm" class="mx-auto relative p-3">
                    <div class="flex items-center space-x-4 rtl:space-x-reverse px-2">
                        <Account class="w-8 h-8 flex-shrink-0 text-gray-500" />
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium text-gray-900 truncate">{user.name}</p>
                        </div>
                        {#if user.id === settings.currentUser.id}
                            <span class="text-xs text-green-600 font-semibold">Active</span>
                        {/if}
                    </div>
                    <button
                        type="button"
                        onclick={() => handleSelectUser(user)}
                        class="absolute inset-0 flex items-center justify-center rounded-lg border -m-[1px] !border-gray-700 {shouldDeleteUser ? 'bg-red-800/85' : 'bg-green-800/85'} opacity-0 hover:opacity-100 cursor-pointer focus:outline-none"
                    >
                        <span class="text-lg font-bold text-gray-100 truncate">
                            {shouldDeleteUser ? 'Delete' : user.id === settings.currentUser.id ? 'Active' : 'Select'}
                        </span>
                        <Indicator
                            class="bg-red-600 hover:bg-red-700"
                            border
                            onmouseenter={() => shouldDeleteUser = true}
                            onmouseleave={() => shouldDeleteUser = false}
                            onclick={() => handleDeleteUser(user.id)}
                            size="xl"
                            placement="top-right"
                        >
                            <span class="text-white text-xs font-bold">—</span>
                        </Indicator>
                    </button>
                </Card>
            </div>
        {/each}
    </div>
    <div class="flex p-0 pt-4 space-x-3">
        <Button onclick={() => { newUserOpen = true; }}>New User</Button>
        <Button onclick={() => usersOpen = false} color="alternative">Close</Button>
    </div>
</Modal>

<Modal title="Add new user" bind:open={newUserOpen} outsideclose>
    <Label class="space-y-2 mb-6">
        <span>Name</span>
        <Input type="text" placeholder="Enter name" bind:value={newUserName} />
    </Label>
    <div class="flex p-0 pt-4 space-x-3">
        <Button onclick={handleAddUser}>Add</Button>
        <Button onclick={() => { newUserOpen = false; newUserName = ''; }} color="alternative">Close</Button>
    </div>
</Modal>

<Footer footerType="logo" class="bg-transparent shadow-none">
    <hr class="my-6 border-gray-200 sm:mx-auto lg:my-8" />
    <FooterCopyright href="https://github.com/tiehfood/thrifty" by="tiehfood" copyrightMessage="| All Rights Reserved"/>
</Footer>
