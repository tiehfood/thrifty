<script lang="ts">
    import "../app.css";
    import debounce from "debounce";
    import type { Flow, Group, User, PageButton } from "$lib/types";
    import { newFlowHandlerStore, editFlowHandlerStore, NUMBER_FORMATS } from "$lib/stores";
    import { sharedState } from "$lib/sharedState.svelte";
    import { settings, loadSettings, patchSettings } from "$lib/settings.svelte";
    import { users, loadUsers, createUser, updateUser, removeUser } from "$lib/users.svelte";
    import { icons, loadIcons, addIcons, removeIcon } from "$lib/icons.svelte";
    import { groups, loadGroups, createGroup, updateGroup, removeGroup } from "$lib/groups.svelte";
    import Fox from "../icons/fox.svg?component";
    import Gear from "../icons/gear.svg?component";
    import Account from "../icons/account.svg?component";
    import {
        Button,
        Card,
        Footer,
        FooterCopyright,
        FooterLink,
        FooterLinkGroup,
        Indicator,
        Input,
        Img,
        Label,
        Modal,
        Navbar,
        NavBrand,
        Checkbox,
        Dropdown,
        DropdownItem,
        Toggle,
    } from "flowbite-svelte";
    import { onMount } from "svelte";

    const editButton: PageButton        = { name: "Edit", clickHandle: clickEdit, color: "light" }
    const newButton: PageButton         = { name: "New Entry", clickHandle: clickNew }
    const groupsButton: PageButton      = { name: "Groups", clickHandle: clickGroups, color: "light" }
    const dummyButton: PageButton       = { name: "", hidden: true }
    const closeButton: PageButton       = { name: "Close", clickHandle: clickClose, color: "light" }
    const modalCloseButton: PageButton  = { name: "Close", clickHandle: clickModalClose, color: "alternative" }
    const modalAddButton: PageButton    = { name: "Add" }
    const modalEditButton: PageButton   = { name: "Edit" }

    let { children } = $props();

    let buttons: PageButton[] = $state([editButton]);

    const PERIOD_OPTIONS = [
        { value: 'weekly',        name: 'weekly' },
        { value: 'bi-weekly',     name: 'bi-weekly' },
        { value: 'bi-monthly',    name: 'bi-monthly' },
        { value: 'tri-monthly',   name: 'tri-monthly' },
        { value: 'semi-annually', name: 'semi-annually' },
        { value: 'annually',      name: 'annually' },
    ];

    const sameWidth = {
        name: 'sameWidth',
        fn({ rects, elements }: { rects: any; elements: any }) {
            Object.assign(elements.floating.style, { width: `${rects.reference.width}px` });
            return {};
        },
    };

    const PERIOD_MULTIPLIERS: Record<string, number> = {
        'weekly':        52 / 12,
        'bi-weekly':     26 / 12,
        'bi-monthly':    1 / 2,
        'tri-monthly':   1 / 3,
        'semi-annually': 1 / 6,
        'annually':      1 / 12,
    };

    let flowModalOpen = $state(false);
    let hiddenFileInputRef: HTMLInputElement;
    let hiddenIconFileInputRef: HTMLInputElement;
    let currentFlow: Flow = $state(getEmptyFlow());
    let notMonthly = $state(false);
    let period = $state('weekly');
    let periodDropdownOpen = $state(false);
    let newFlowHandler: (flow: Flow) => void;
    newFlowHandlerStore.subscribe((handler: (flow: Flow) => void) => newFlowHandler = handler);

    let settingsOpen = $state(false);
    let settingsLeftColHeight = $state(0);

    let galleryOpen = $state(false);
    let galleryOnSelect: (data: string) => void = $state(() => {});

    async function openGallery(onSelect: (data: string) => void) {
        await loadIcons();
        galleryOnSelect = onSelect;
        galleryOpen = true;
    }

    let groupsOpen = $state(false);
    let newGroupOpen = $state(false);
    let shouldDeleteGroup = $state(false);
    let currentGroup: Group = $state(getEmptyGroup());
    let hiddenGroupFileInputRef: HTMLInputElement;
    let groupDropdownOpen = $state(false);

    function getEmptyGroup(): Group {
        return { id: undefined, name: "", description: "", icon: undefined, amount: 0 };
    }

    async function clickGroups() {
        await loadGroups();
        groupsOpen = true;
    }

    async function handleGroupSubmit() {
        if (!currentGroup.name.trim()) {
            alert("Please enter a name.");
            return;
        }
        if (currentGroup.id) {
            await updateGroup(currentGroup.id, currentGroup);
        } else {
            await createGroup(currentGroup);
        }
        sharedState.reloadTrigger++;
        newGroupOpen = false;
    }

    async function handleDeleteGroup(id: string) {
        await removeGroup(id);
        sharedState.reloadTrigger++;
        shouldDeleteGroup = false;
        if (groups.length === 0) groupsOpen = false;
    }

    function handleGroupFileUpload(event: Event) {
        const file = (event.target as HTMLInputElement).files?.[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (e: ProgressEvent) => {
                const svg = (e.target as FileReader).result ?? '';
                currentGroup.icon = `data:image/svg+xml;base64,${btoa(svg.toString())}`;
            };
            reader.readAsText(file);
        }
        (event.target as HTMLInputElement).value = '';
    }
    let iconContainerMaxH = $derived(
        settingsLeftColHeight > 0 ? Math.max(0, settingsLeftColHeight - 96) : 272
    );

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
        buttons = [dummyButton, newButton, groupsButton, closeButton];
        sharedState.isEditMode = true;
    }

    async function clickNew() {
        currentFlow = getEmptyFlow();
        notMonthly = false;
        period = 'weekly';
        await loadGroups();
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
        if (notMonthly) {
            currentFlow.amount = Math.round(currentFlow.amount * (PERIOD_MULTIPLIERS[period] ?? 1) * 100) / 100;
        }
        const errors = validateForm(currentFlow);
        if (errors.length > 0) {
            alert(errors.join("\n"));
            return;
        }
        if (newFlowHandler) {
            await newFlowHandler(currentFlow);
        }
        sharedState.reloadTrigger++;
        flowModalOpen = false;
        currentFlow = getEmptyFlow();
        notMonthly = false;
        period = 'weekly';
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

    function handleIconUpload(event: Event) {
        const input = event.target as HTMLInputElement;
        const files = Array.from(input.files ?? []);
        const reads = files.map(file => new Promise<string>(resolve => {
            const reader = new FileReader();
            reader.onload = (e: ProgressEvent) => {
                const svgContent = (e.target as FileReader).result ?? '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><path d="m0,0v1h1V0"/></svg>';
                resolve(`data:image/svg+xml;base64,${btoa(svgContent.toString())}`);
            };
            reader.readAsText(file);
        }));
        Promise.all(reads).then(dataUris => addIcons(dataUris));
        input.value = '';
    }

    async function editFlowHandler(flow: Flow) {
        currentFlow = flow;
        await loadGroups();
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
        <Button onclick={() => { loadIcons(); settingsOpen = true; }} color="light" class="me-3 px-3 py-2 sm:py-2.5">
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
        <div class="grid grid-cols-2 gap-8 mb-6">
            <Label class="space-y-2">
                <span>Description</span>
                <Input type="text" placeholder="Enter description (optional)" bind:value={currentFlow.description}/>
            </Label>
            <Label class="space-y-2">
                <span>Group (optional)</span>
                {#if groups.length === 0}
                    <p class="text-sm text-gray-500 py-2">No groups</p>
                {:else}
                    <Button type="button" color="light" class="w-full flex justify-between items-center px-2.5!">
                        {groups.find(g => g.id === currentFlow.groupId)?.name ?? 'None'}
                        <svg class="w-2.5 h-2.5 ms-3" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 10 6">
                            <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 1 4 4 4-4"/>
                        </svg>
                    </Button>
                    <Dropdown simple bind:isOpen={groupDropdownOpen} middlewares={[sameWidth]}>
                        <DropdownItem
                            onclick={() => { currentFlow.groupId = undefined; groupDropdownOpen = false; }}
                            class={!currentFlow.groupId ? 'font-semibold text-primary-600' : ''}
                            style="padding-left: 0.625rem; padding-right: 0.625rem;"
                        >None</DropdownItem>
                        {#each groups as group}
                            <DropdownItem
                                onclick={() => { currentFlow.groupId = group.id; groupDropdownOpen = false; }}
                                class={currentFlow.groupId === group.id ? 'font-semibold text-primary-600' : ''}
                                style="padding-left: 0.625rem; padding-right: 0.625rem;"
                            >{group.name}</DropdownItem>
                        {/each}
                    </Dropdown>
                {/if}
            </Label>
        </div>
        <div class="grid grid-cols-2 gap-8">
            <div class="space-y-2 mb-6">
                <Label for="numberInput">Amount</Label>
                <div class="flex items-center gap-3">
                    <div class="w-28">
                        <Input type="number" onfocus={handleFocus} id="numberInput" step="0.01" bind:value={currentFlow.amount}/>
                    </div>
                    {#if !currentFlow.id}
                        <Checkbox bind:checked={notMonthly} class="accent-primary-700">not monthly</Checkbox>
                    {/if}
                </div>
                {#if !currentFlow.id && notMonthly}
                    <Button type="button" color="light" class="w-full flex justify-between items-center px-2.5!">
                        {PERIOD_OPTIONS.find(o => o.value === period)?.name}
                        <svg class="w-2.5 h-2.5 ms-3" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 10 6">
                            <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 1 4 4 4-4"/>
                        </svg>
                    </Button>
                    <Dropdown simple bind:isOpen={periodDropdownOpen} middlewares={[sameWidth]}>
                        {#each PERIOD_OPTIONS as opt}
                            <DropdownItem
                                onclick={() => { period = opt.value; periodDropdownOpen = false; }}
                                class={period === opt.value ? 'font-semibold text-primary-600' : ''}
                                style="padding-left: 0.625rem; padding-right: 0.625rem;"
                            >{opt.name}</DropdownItem>
                        {/each}
                    </Dropdown>
                {/if}
            </div>
            <div class="flex items-center justify-center gap-2 self-start">
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
                <Button type="button" color="alternative" onclick={() => openGallery((data) => { currentFlow.icon = data; })}>Gallery</Button>
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

<Modal title="Choose icon" bind:open={galleryOpen} outsideclose size="sm" class="!w-fit">
    {#if icons.length === 0}
        <div class="flex items-center justify-center rounded p-4 ring-2 ring-gray-300 bg-gray-100">
            <span class="text-center text-gray-600">Add icons on settings page</span>
        </div>
    {:else}
        <div class="p-2 ring-2 ring-gray-300 bg-gray-100 rounded w-fit max-h-[188px] overflow-y-auto">
            <div class="grid grid-cols-8 gap-1">
                {#each icons as icon}
                    <button
                        type="button"
                        onclick={() => { galleryOnSelect(icon.data); galleryOpen = false; }}
                        class="relative h-10 w-10 flex-shrink-0 rounded cursor-pointer hover:ring-2 hover:ring-primary-400 {currentFlow.icon === icon.data ? 'ring-2 ring-primary-500' : ''}"
                    >
                        <Img class="object-contain w-full h-full" src={icon.data} alt="Icon" />
                    </button>
                {/each}
            </div>
        </div>
    {/if}
    <div class="flex p-0 pt-4">
        <Button onclick={() => galleryOpen = false} color="alternative">Close</Button>
    </div>
</Modal>

<Modal title="Groups" bind:open={groupsOpen} outsideclose>
    <div>
        {#each groups as group}
            <div class="p-1">
                <Card size="sm" class="mx-auto relative p-3">
                    <div class="flex items-center space-x-4 rtl:space-x-reverse px-2">
                        <Img class="justify-center rounded-none w-10 h-10 flex-shrink-0 bg-white object-contain" src={group.icon} alt="Icon" />
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium text-gray-900 truncate">{group.name}</p>
                            <p class="text-xs text-gray-500 truncate">{group.description}</p>
                        </div>
                    </div>
                    <button
                        type="button"
                        onclick={() => { if (!shouldDeleteGroup) { currentGroup = { ...group }; newGroupOpen = true; } }}
                        class="absolute inset-0 flex items-center justify-center rounded-lg border -m-[1px] !border-gray-700 {shouldDeleteGroup ? 'bg-red-800/85' : 'bg-green-800/85'} opacity-0 hover:opacity-100 cursor-pointer focus:outline-none"
                    >
                        <span class="text-lg font-bold text-gray-100 truncate">{shouldDeleteGroup ? 'Delete' : 'Edit'}</span>
                        <Indicator
                            class="bg-red-600 hover:bg-red-700"
                            border
                            onmouseenter={() => shouldDeleteGroup = true}
                            onmouseleave={() => shouldDeleteGroup = false}
                            onclick={() => handleDeleteGroup(group.id as string)}
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
        <Button onclick={() => { currentGroup = getEmptyGroup(); newGroupOpen = true; }}>New Group</Button>
        <Button onclick={() => groupsOpen = false} color="alternative">Close</Button>
    </div>
</Modal>

<Modal title={currentGroup.id ? "Edit group" : "Add new group"} bind:open={newGroupOpen} outsideclose>
    <Label class="space-y-2 mb-6">
        <span>Name</span>
        <Input type="text" placeholder="Enter name" bind:value={currentGroup.name}/>
    </Label>
    <Label class="space-y-2 mb-6">
        <span>Description</span>
        <Input type="text" placeholder="Enter description (optional)" bind:value={currentGroup.description}/>
    </Label>
    <div class="flex items-center justify-center gap-2 self-start mb-6">
        <div class="flex justify-center">
            <input type="file" accept=".svg" onchange={handleGroupFileUpload} class="hidden" bind:this={hiddenGroupFileInputRef} />
            <button type="button" onclick={() => hiddenGroupFileInputRef.click()} class="relative flex {currentGroup.icon === undefined ? 'items-center' : ''} justify-center rounded p-1 ring-2 ring-gray-300 {currentGroup.icon === undefined ? 'bg-gray-100' : 'bg-white'} aspect-square h-24 w-24 m-1 flex-shrink-0 cursor-pointer">
                {#if currentGroup.icon === undefined}
                    <span class="text-center text-gray-600">Icon (optional)</span>
                {:else}
                    <Img class="object-contain w-full h-full" src={currentGroup.icon} alt="Icon" />
                {/if}
            </button>
        </div>
        <Button type="button" color="alternative" onclick={() => openGallery((data) => { currentGroup.icon = data; })}>Gallery</Button>
    </div>
    <div class="flex p-0 pt-4 space-x-3">
        <Button onclick={handleGroupSubmit}>{currentGroup.id ? 'Edit' : 'Add'}</Button>
        <Button onclick={() => newGroupOpen = false} color="alternative">Close</Button>
    </div>
</Modal>

<Modal title="Settings" bind:open={settingsOpen} outsideclose size="md">
    <div class="grid grid-cols-2 gap-4">
        <div class="flex flex-col" bind:clientHeight={settingsLeftColHeight}>
            <Label class="space-y-2 mb-6">
                <Toggle bind:checked={multiUserEnabledLocal} onchange={handleMultiUserToggle}>
                    Multi user support
                </Toggle>
            </Label>
            <div class="mb-6">
                <span class="block mb-2 text-sm font-medium text-gray-900">Number format</span>
                <div class="space-y-3">
                    {#each NUMBER_FORMATS as fmt}
                        <div class="flex items-center gap-2">
                            <input
                                type="radio"
                                id="fmt-{fmt.value}"
                                name="numberFormat"
                                value={fmt.value}
                                bind:group={settings.numberFormat}
                                onchange={() => patchSettings({ numberFormat: fmt.value })}
                                class="w-5 h-5 cursor-pointer accent-primary-600"
                            />
                            <label for="fmt-{fmt.value}" class="text-sm font-medium text-gray-900 cursor-pointer">{fmt.name}</label>
                        </div>
                    {/each}
                </div>
            </div>
            <div class="mt-auto pt-4">
                <Button onclick={() => settingsOpen = false} color="alternative">Close</Button>
            </div>
        </div>
        <div class="flex flex-col">
            <div class="flex items-center min-h-6 mb-6">
                <span class="text-sm font-medium text-gray-900">Icons</span>
            </div>
            <input type="file" accept=".svg" multiple onchange={handleIconUpload} class="hidden" bind:this={hiddenIconFileInputRef} />
            <div class="overflow-y-auto rounded p-2 ring-2 ring-gray-300 bg-gray-100" style="max-height: {iconContainerMaxH}px">
                {#if icons.length === 0}
                    <div class="flex items-center justify-center p-2">
                        <span class="text-center text-gray-600">No icons</span>
                    </div>
                {/if}
                <div class="grid grid-cols-6 gap-1">
                    {#each icons as icon}
                        <div class="relative aspect-square">
                            <Img class="object-contain w-full h-full" src={icon.data} alt="Icon" />
                            <button
                                type="button"
                                onclick={() => { if (!icon.isUsed) removeIcon(icon.id); }}
                                class="absolute inset-0 flex items-center justify-center rounded border -m-[1px] !border-gray-700 text-xs font-semibold text-white {icon.isUsed ? 'bg-green-800/85 cursor-default' : 'bg-red-800/85 cursor-pointer'} opacity-0 hover:opacity-100"
                            >
                                {icon.isUsed ? 'used' : 'delete'}
                            </button>
                        </div>
                    {/each}
                </div>
            </div>
            <div class="pt-2">
                <Button onclick={() => hiddenIconFileInputRef.click()}>Add Icon</Button>
            </div>
        </div>
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
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-center text-center sm:text-left">
    <FooterCopyright href="https://buymeacoffee.com/tiehfood" by="tiehfood" copyrightMessage="| All Rights Reserved |&nbsp;"/>
        <FooterLinkGroup class="flex flex-wrap items-center text-sm text-gray-500 sm:mt-0">
            {#if VERSION !== "dev"}
                <FooterLink href="https://github.com/tiehfood/thrifty/releases/tag/v{VERSION}">{VERSION}</FooterLink>
            {:else}
                <FooterLink href="https://github.com/tiehfood/thrifty">dev</FooterLink>
            {/if}
        </FooterLinkGroup>
    </div>
</Footer>
