const protocol = window.location.protocol.replace(':', '');
const hostname = window.location.hostname.toLowerCase();
const port = 8081; //window.location.port.trim();

export const sharedState = $state({ isEditMode: false, multiUserEnabled: false, reloadTrigger: 0, apiBase: `${protocol}://${hostname}${port ? ':' + port : ''}` });
