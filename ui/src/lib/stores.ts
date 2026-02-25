import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Flow, NumberFormatOption } from "$lib/types";

export const newFlowHandlerStore: Writable<(flow: Flow) => void> = writable();
export const editFlowHandlerStore: Writable<(flow: Flow) => void> = writable();

export const NUMBER_FORMATS: NumberFormatOption[] = [
    { value: 'eu-decimal',     name: '#.###(,##)',  locales: 'de-DE', format: { style: 'currency', currency: 'EUR', trailingZeroDisplay: 'stripIfInteger' } },
    { value: 'eu-decimal-fix', name: '#.###,##',    locales: 'de-DE', format: { style: 'currency', currency: 'EUR', minimumFractionDigits: 2, maximumFractionDigits: 2 } },
    { value: 'us-decimal',     name: '#,###(.##)',  locales: 'en-US', format: { style: 'currency', currency: 'EUR', trailingZeroDisplay: 'stripIfInteger' } },
    { value: 'us-decimal-fix', name: '#,###.##',    locales: 'en-US', format: { style: 'currency', currency: 'EUR', minimumFractionDigits: 2, maximumFractionDigits: 2 } },
    { value: 'eu-integer',     name: '#.###',       locales: 'de-DE', format: { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 } },
    { value: 'us-integer',     name: '#,###',       locales: 'en-US', format: { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 } },
    { value: 'compact',        name: '#k',          locales: 'en-US', format: { notation: 'compact', compactDisplay: 'short' } },
];
