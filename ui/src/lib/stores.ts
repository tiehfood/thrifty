import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Flow } from "$lib/types";

export const newFlowHandlerStore: Writable<(flow: Flow) => void> = writable();
export const editFlowHandlerStore: Writable<(flow: Flow) => void> = writable();
