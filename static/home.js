// home.js

// Initialize state variables
let activeRequests = 0; // Counter for active requests
let isShowingUntrusted = false; // Default view

// Assume these lists are declared and initialized globally elsewhere
// If not, uncomment the following lines to initialize them
// let TrustedCharacters = TrustedCharacters || [];
// let TrustedCorporations = TrustedCorporations || [];
// let UntrustedCharacters = UntrustedCharacters || [];
// let UntrustedCorporations = UntrustedCorporations || [];

// Initialize table variables grouped under a single object
const tables = {};

// Configure Toastr options for notifications
toastr.options = {
    "closeButton": true,
    "debug": false,
    "newestOnTop": true,
    "progressBar": true,
    "positionClass": "toast-top-right",
    "preventDuplicates": false,
    "onclick": null,
    "showDuration": "300",
    "hideDuration": "1000",
    "timeOut": "1500", // Increased timeout for better visibility
    "extendedTimeOut": "1000",
    "showEasing": "swing",
    "hideEasing": "linear",
    "showMethod": "fadeIn",
    "hideMethod": "fadeOut"
};

/**
 * Function to show the loading indicator
 */
function showLoading() {
    activeRequests += 1;
    const loader = document.getElementById("loading-indicator");
    if (loader) {
        loader.hidden = false;
    }
}

/**
 * Function to hide the loading indicator
 */
function hideLoading() {
    if (activeRequests > 0) {
        activeRequests -= 1;
    }
    if (activeRequests === 0) {
        const loader = document.getElementById("loading-indicator");
        if (loader) {
            loader.hidden = true;
        }
    }
}

/**
 * Function to toggle the disabled state of a button
 * @param {HTMLElement} button - The button element to toggle
 * @param {boolean} disable - Whether to disable or enable the button
 */
function toggleButtonState(button, disable) {
    if (!button) return;

    button.disabled = disable;
    button.style.cursor = disable ? "not-allowed" : "pointer";
    button.style.opacity = disable ? "0.6" : "1";
}

/**
 * Function to resize a Tabulator table container
 * @param {string} tableId - The ID of the table container
 */
function resizeTabulatorTable(tableId) {
    const tableInstance = tables[tableId];
    if (tableInstance) {
        tableInstance.redraw(true);
    }
}

/**
 * Function to toggle the display of a section
 * @param {string} sectionId - The ID of the section to toggle
 * @param {boolean} show - Whether to show or hide the section
 */
function toggleSectionDisplay(sectionId, show) {
    const section = document.getElementById(sectionId);
    if (!section) {
        console.error(`Section with ID "${sectionId}" not found.`);
        return;
    }
    section.style.display = show ? "block" : "none";
}

/**
 * Function to toggle multiple sections
 * @param {Array} sections - Array of section IDs to toggle
 * @param {boolean} show - Whether to show or hide the sections
 */
function toggleMultipleSections(sections, show) {
    sections.forEach(sectionId => toggleSectionDisplay(sectionId, show));
}

/**
 * Helper function to determine the correct server endpoint
 * @param {string} trustStatus - 'trusted' or 'untrusted'
 * @param {string} entityType - 'character' or 'corporation'
 * @param {string} action - 'add' or 'remove'
 * @returns {string|null} - The server endpoint URL or null if invalid inputs
 */
function getServerEndpoint(trustStatus, entityType, action) {
    const endpoints = {
        trusted: {
            character: {
                add: '/validate-and-add-trusted-character',
                remove: '/remove-trusted-character',
            },
            corporation: {
                add: '/validate-and-add-trusted-corporation',
                remove: '/remove-trusted-corporation',
            }
        },
        untrusted: {
            character: {
                add: '/validate-and-add-untrusted-character',
                remove: '/remove-untrusted-character',
            },
            corporation: {
                add: '/validate-and-add-untrusted-corporation',
                remove: '/remove-untrusted-corporation',
            }
        }
    };

    return endpoints[trustStatus]?.[entityType]?.[action] || null;
}

/**
 * Centralized fetch function that handles JSON and text responses.
 * Throws an error with the appropriate message based on response status.
 * @param {string} url - The endpoint URL.
 * @param {object} options - Fetch options.
 * @returns {Promise<Object|string>} - Parsed JSON object or plain text string.
 */
async function fetchWithHandling(url, options) {
    try {
        const response = await fetch(url, options);
        const contentType = response.headers.get("Content-Type");

        if (!response.ok) {
            let errorData;
            if (contentType && contentType.includes("application/json")) {
                errorData = await response.json();
                throw new Error(errorData.error || "An error occurred.");
            } else {
                errorData = await response.text();
                throw new Error(errorData || "An error occurred.");
            }
        }

        if (contentType && contentType.includes("application/json")) {
            return response.json();
        } else {
            return response.text();
        }
    } catch (error) {
        console.error(`Network or parsing error: ${error}`);
        throw error;
    }
}

/**
 * Function to write contacts
 * Calls /add-contacts and /delete-contacts endpoints sequentially
 * @param {number} characterID - ID of the character
 */
async function writeContacts(characterID) {
    // Validate characterID
    if (typeof characterID !== 'number' || isNaN(characterID) || characterID <= 0) {
        toastr.error("Invalid Character ID.");
        console.error("Invalid Character ID:", characterID);
        return;
    }

    showLoading();
    const toggleBtn = document.getElementById("toggle-contacts-btn");
    toggleButtonState(toggleBtn, true);

    try {
        console.log("Sending to /add-contacts:", { characterID });

        // Call /add-contacts endpoint
        let response = await fetch(`/add-contacts`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ characterID })
        });

        if (!response.ok) {
            let errorMessage = "Failed to add contacts.";
            try {
                const errData = await response.json();
                errorMessage = errData.error || errorMessage;
            } catch (e) {
                const errText = await response.text();
                errorMessage = errText || errorMessage;
            }
            throw new Error(errorMessage);
        }

        const addData = await response.json();
        console.log("Contacts added successfully:", addData);

        console.log("Sending to /delete-contacts:", { characterID });

        // Call /delete-contacts endpoint
        response = await fetch(`/delete-contacts`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ characterID })
        });

        if (!response.ok) {
            let errorMessage = "Failed to delete contacts.";
            try {
                const errData = await response.json();
                errorMessage = errData.error || errorMessage;
            } catch (e) {
                const errText = await response.text();
                errorMessage = errText || errorMessage;
            }
            throw new Error(errorMessage);
        }

        const deleteData = await response.json();
        console.log("Contacts deleted successfully:", deleteData);
        toastr.success("Contacts updated successfully.");
    } catch (error) {
        toastr.error("Error writing contacts: " + error.message);
        console.error("Error writing contacts:", error);
    } finally {
        hideLoading();
        toggleButtonState(toggleBtn, false);
    }
}



/**
 * Checks if an entity is in a given list by Identifier (ID or Name).
 * @param {Array} list - The list to check.
 * @param {string|number} identifier - The ID or Name of the entity.
 * @param {string} entityType - 'character' or 'corporation'.
 * @returns {boolean} - True if the entity is in the list.
 */
function isEntityInListByIdentifier(list, identifier, entityType) {
    if (!identifier) {
        console.warn(`Invalid identifier:`, identifier);
        return false;
    }

    const isId = typeof identifier === 'number' || /^\d+$/.test(identifier);
    const idField = entityType === 'character' ? 'CharacterID' : 'CorporationID';
    const nameField = entityType === 'character' ? 'CharacterName' : 'CorporationName';

    if (isId) {
        const numericId = typeof identifier === 'number' ? identifier : parseInt(identifier, 10);
        if (isNaN(numericId)) {
            console.warn(`Identifier "${identifier}" is not a valid number.`);
            return false;
        }
        return list.some(entity => entity[idField] === numericId);
    } else {
        if (typeof identifier !== 'string') {
            console.warn(`Identifier must be a string when not numeric.`);
            return false;
        }
        return list.some(entity => {
            const entityName = entity[nameField];
            if (typeof entityName !== 'string') {
                console.warn(`Invalid ${nameField} in entity:`, entity);
                return false;
            }
            return entityName.toLowerCase() === identifier.toLowerCase();
        });
    }
}

/** Helper functions for character and corporation checks */
const isCharacterInTrustedList = (identifier) => isEntityInListByIdentifier(TrustedCharacters, identifier, 'character');
const isCorporationInTrustedList = (identifier) => isEntityInListByIdentifier(TrustedCorporations, identifier, 'corporation');
const isCharacterInUntrustedList = (identifier) => isEntityInListByIdentifier(UntrustedCharacters, identifier, 'character');
const isCorporationInUntrustedList = (identifier) => isEntityInListByIdentifier(UntrustedCorporations, identifier, 'corporation');

/** Helper functions for ID-based checks */
const isCharacterInTrustedListByID = (id) => isEntityInListByIdentifier(TrustedCharacters, id, 'character');
const isCorporationInTrustedListByID = (id) => isEntityInListByIdentifier(TrustedCorporations, id, 'corporation');

/**
 * Checks if a character is trusted based on character and corporation trust lists.
 * @param {object} character - The character object.
 * @returns {boolean} - True if trusted, otherwise false.
 */
/**
 * Checks if a character is trusted based on character and corporation trust lists.
 * @param {object} character - The character object.
 * @returns {boolean} - True if trusted, otherwise false.
 */
function isCharacterTrusted(character) {
    return isCharacterInTrustedListByID(character.CharacterID) || isCorporationInTrustedListByID(character.CorporationID);
}


function updateTileTrustStatus(tile, isTrusted) {
    console.log(`Updating tile trust status. Current class: ${tile.className}, New status: ${isTrusted ? 'Trusted' : 'Untrusted'}`);

    // Define the possible states
    const classMap = {
        trusted: "trusted",
        untrusted: "untrusted"
    };

    // Remove all trust-related classes before applying new ones
    tile.classList.remove(classMap.trusted, classMap.untrusted);

    // Apply the appropriate class
    if (isTrusted) {
        tile.classList.add(classMap.trusted); // Teal border for trusted
        console.log(`Applied 'trusted' class to tile.`);
    } else {
        tile.classList.add(classMap.untrusted); // Yellow border for untrusted
        console.log(`Applied 'untrusted' class to tile.`);
    }
}


/**
 * Recomputes and updates a character's tile trust status.
 * @param {number|string} characterID - The ID of the character.
 */
function recomputeAndUpdateTileTrustStatus(characterID) {
    // Convert characterID to number to match the data type in TabulatorIdentities
    const numericCharacterID = Number(characterID);
    if (isNaN(numericCharacterID)) {
        console.warn(`Invalid character ID: ${characterID}`);
        return;
    }

    const character = TabulatorIdentities.find(char => char.CharacterID === numericCharacterID);
    if (character) {
        const tile = document.querySelector(`.character-tile[data-id="${character.CharacterID}"]`);
        if (tile) {
            const isTrusted = isCharacterTrusted(character);
            updateTileTrustStatus(tile, isTrusted);
            console.log(`Updated trust status for CharacterID ${character.CharacterID}: ${isTrusted ? 'Trusted' : 'Untrusted'}`);
        }
    } else {
        console.warn(`Character with ID ${characterID} not found.`);
    }
}

/**
 * Initializes a character tile
 * @param {object} character - The character data object
 * @returns {HTMLElement} - The character tile element
 */
function initializeCharacterTile(character) {
    const tile = document.createElement("div");
    tile.className = "character-tile";
    tile.dataset.id = character.CharacterID; // Ensure correct property name

    // Determine if the character is trusted based on both character and corporation trust lists
    const isTrusted = isCharacterTrusted(character);
    updateTileTrustStatus(tile, isTrusted);

    const img = document.createElement("img");
    img.src = character.Portrait;
    img.alt = character.CharacterName;
    img.className = "character-portrait";

    const name = document.createElement("div");
    name.className = "character-name";
    name.innerText = character.CharacterName;

    const button = document.createElement("button");
    button.className = "write-contacts-btn";
    button.title = "Write Contacts";
    button.setAttribute("data-tooltip", "Write Contacts");
    button.setAttribute("aria-label", "Write Contacts");
    button.innerHTML = '<i class="fas fa-pen" aria-hidden="true"></i>';

    button.addEventListener("click", (e) => {
        e.stopPropagation();
        writeContacts(character.CharacterID);
    });

    tile.appendChild(img);
    tile.appendChild(name);
    tile.appendChild(button);
    return tile;
}

/**
 * Initializes all character tiles
 */
function initializeCharacterTiles() {
    const characterContainer = document.getElementById("character-container");
    if (!characterContainer) {
        console.error(`Character container with ID "character-container" not found.`);
        return;
    }
    characterContainer.innerHTML = ""; // Clear existing tiles to prevent duplicates
    characterContainer.style.display = "flex";
    characterContainer.style.flexWrap = "wrap";
    characterContainer.style.justifyContent = "center";
    characterContainer.style.gap = "20px";

    TabulatorIdentities.forEach(character => {
        const tile = initializeCharacterTile(character);
        characterContainer.appendChild(tile);

        // Click event to add untrusted characters to the trusted list
        tile.addEventListener("click", () => {
            console.log(`Tile clicked for CharacterID: ${character.CharacterID}, Class: ${tile.className}`);
            if (tile.classList.contains('untrusted') && activeRequests === 0) {
                console.log("Adding to trusted list...");
                addEntity('trusted', 'character', character.CharacterID.toString()); // Convert to string
            } else {
                console.log("Click ignored. Either trusted or request in progress.");
            }
        });
    });
}

/**
 * Adds an entity based on trustStatus and entityType using a single identifier.
 * @param {string} trustStatus - 'trusted' or 'untrusted'.
 * @param {string} entityType - 'character' or 'corporation'.
 * @param {string|number} identifier - Character/Corporation ID or Name.
 */
function addEntity(trustStatus, entityType, identifier) {
    console.log("Adding entity:", trustStatus, entityType, identifier);

    const serverEndpoint = getServerEndpoint(trustStatus, entityType, 'add');
    if (!serverEndpoint) {
        console.error("Invalid trustStatus or entityType provided to addEntity.");
        toastr.error("An unexpected error occurred.");
        return;
    }

    // Ensure identifier is always a string
    const identifierStr = String(identifier);

    const oppositeTrustStatus = trustStatus === 'trusted' ? 'untrusted' : 'trusted';

    // Check if the entity is already in the opposite list
    const isInOppositeList = trustStatus === 'trusted'
        ? (entityType === 'character' ? isCharacterInUntrustedList(identifier) : isCorporationInUntrustedList(identifier))
        : (entityType === 'character' ? isCharacterInTrustedList(identifier) : isCorporationInTrustedList(identifier));

    if (isInOppositeList) {
        toastr.warning(`${capitalize(entityType)} already exists in the ${oppositeTrustStatus} list.`);
        console.warn(`${capitalize(entityType)} with identifier ${identifierStr} is already in ${oppositeTrustStatus} list.`);
        return; // Prevent adding to the current list
    }

    // Prepare payload with a single identifier field
    const payload = {
        identifier: identifierStr
    };

    showLoading();
    const toggleBtn = document.getElementById("toggle-contacts-btn");
    toggleButtonState(toggleBtn, true);

    fetchWithHandling(serverEndpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    })
        .then(data => {
            console.log(`Entity added:`, data);

            // Update local data and the table
            additionUpdateLocalData(trustStatus, entityType, data);
            const tableId = `${trustStatus}-${entityType}s-table`;
            addRowToTable(tableId, data);

            // Recompute and update the trust status of affected characters
            if (entityType === 'character') {
                recomputeAndUpdateTileTrustStatus(data.CharacterID);
            } else if (entityType === 'corporation') {
                // Update all characters belonging to this corporation
                TabulatorIdentities
                    .filter(char => char.CorporationID === data.CorporationID)
                    .forEach(char => recomputeAndUpdateTileTrustStatus(char.CharacterID));
            }

        })
        .catch(error => {
            console.error(`Error adding entity: ${error}`);
            toastr.error(`Failed to add ${entityType}. ${error.message}`);
        })
        .finally(() => {
            hideLoading();
            toggleButtonState(toggleBtn, false);
        });
}


/**
 * Updates local data by adding the new entity to the appropriate list
 * @param {string} trustStatus - 'trusted' or 'untrusted'
 * @param {string} entityType - 'character' or 'corporation'
 * @param {object} data - The data object of the entity to add
 */
function additionUpdateLocalData(trustStatus, entityType, data) {
    let targetList;
    if (trustStatus === 'trusted' && entityType === 'character') {
        targetList = TrustedCharacters;
    } else if (trustStatus === 'trusted' && entityType === 'corporation') {
        targetList = TrustedCorporations;
    } else if (trustStatus === 'untrusted' && entityType === 'character') {
        targetList = UntrustedCharacters;
    } else if (trustStatus === 'untrusted' && entityType === 'corporation') {
        targetList = UntrustedCorporations;
    } else {
        console.warn(`Unknown trustStatus (${trustStatus}) or entityType (${entityType})`);
        return;
    }

    // Check for duplicates
    const idField = entityType === 'character' ? 'CharacterID' : 'CorporationID';
    const exists = targetList.some(entity => entity[idField] === data[idField]);

    if (!exists) {
        targetList.push(data);
        console.log(`Added ${entityType} with ID ${data[idField]} to ${trustStatus} list.`);
    } else {
        console.warn(`${entityType} with ID ${data[idField]} already exists in ${trustStatus} list.`);
    }
}


/**
 * Adds a row to the specified table
 * @param {string} tableId - The ID of the table
 * @param {object} data - The data object to add as a row
 */
function addRowToTable(tableId, data) {
    const targetTable = tables[tableId];
    if (targetTable) {
        const rowID = data.CharacterID || data.CorporationID;
        const existingRow = targetTable.getRow(rowID);

        if (!existingRow) {
            targetTable.addRow(data)
                .then(() => {
                    const entityType = tableId.includes('character') ? 'Character' : 'Corporation';
                    const trustStatus = tableId.split('-')[0];
                    toastr.success(`Added ${data[`${entityType}Name`]} to ${capitalize(trustStatus)} list.`);

                    // Determine if the table should be displayed based on current view
                    const isTrustedTable = tableId.startsWith('trusted-');
                    const isUntrustedTable = tableId.startsWith('untrusted-');

                    if (isTrustedTable && !isShowingUntrusted) {
                        // Trusted tables should be visible when showing trusted view
                        toggleSectionDisplay(tableId, true);
                    } else if (isUntrustedTable && isShowingUntrusted) {
                        // Untrusted tables should be visible when showing untrusted view
                        toggleSectionDisplay(tableId, true);
                    } else {
                        // Tables not in the current view should remain hidden
                        toggleSectionDisplay(tableId, false);
                    }

                    resizeTabulatorTable(tableId);
                })
                .catch(error => {
                    console.error(`Error adding row to ${tableId}:`, error);
                    const entityType = tableId.includes('character') ? 'Character' : 'Corporation';
                    const trustStatus = tableId.split('-')[0];
                    toastr.error(`Error updating ${capitalize(trustStatus)} ${entityType} table.`);
                });
        } else {
            const entityType = tableId.includes('character') ? 'Character' : 'Corporation';
            const trustStatus = tableId.split('-')[0];
            toastr.warning(`${entityType} already exists in the ${capitalize(trustStatus)} list.`);
        }
    }
}

/**
 * Handle form submission by extracting the identifier and performing add/remove operations.
 * @param {Event} e - The submit event.
 * @param {string} trustStatus - 'trusted' or 'untrusted'.
 * @param {string} entityType - 'character' or 'corporation'.
 */
function handleFormSubmission(e, trustStatus, entityType) {
    e.preventDefault();

    const inputIdMap = {
        'trusted-character': 'trusted-character-identifier',
        'trusted-corporation': 'trusted-corporation-identifier',
        'untrusted-character': 'untrusted-character-identifier',
        'untrusted-corporation': 'untrusted-corporation-identifier',
    };

    const inputId = inputIdMap[`${trustStatus}-${entityType}`];
    const inputElement = document.getElementById(inputId);
    if (!inputElement) {
        toastr.error(`Input element with ID "${inputId}" not found.`);
        console.error(`Input element with ID "${inputId}" not found.`);
        return;
    }

    const identifier = inputElement.value.trim();

    console.log(`Form Submission - Trust Status: ${trustStatus}, Entity Type: ${entityType}, Identifier: "${identifier}"`);

    if (!identifier) {
        toastr.error(`${capitalize(entityType)} needs a name or id.`);
        return;
    }

    // Optional: Additional validation based on expected formats
    if (isNumeric(identifier) && parseInt(identifier, 10) <= 0) {
        toastr.error("Identifier must be a positive number.");
        return;
    }

    // Prevent adding an entity that already exists in the opposite list
    const isInOppositeList = trustStatus === 'trusted'
        ? (entityType === 'character' ? isCharacterInUntrustedList(identifier) : isCorporationInUntrustedList(identifier))
        : (entityType === 'character' ? isCharacterInTrustedList(identifier) : isCorporationInTrustedList(identifier));

    if (isInOppositeList) {
        const statusMessage = trustStatus === 'trusted' ? 'untrusted' : 'trusted';
        toastr.warning(`${identifier} is already in the ${statusMessage} list.`);
        return;
    }

    // Call addEntity with the single identifier
    addEntity(trustStatus, entityType, identifier);
    inputElement.value = '';  // Clear the input field after submission
}

/**
 * Setup event listeners for forms using a centralized handler
 */
function setupFormEventListeners() {
    const forms = [
        { id: "add-trusted-character-form", trustStatus: 'trusted', entityType: 'character' },
        { id: "add-untrusted-character-form", trustStatus: 'untrusted', entityType: 'character' },
        { id: "add-trusted-corporation-form", trustStatus: 'trusted', entityType: 'corporation' },
        { id: "add-untrusted-corporation-form", trustStatus: 'untrusted', entityType: 'corporation' },
    ];

    forms.forEach(({ id, trustStatus, entityType }) => {
        const form = document.getElementById(id);
        if (form) {
            form.addEventListener("submit", (e) => handleFormSubmission(e, trustStatus, entityType));
        } else {
            console.warn(`Form with ID "${id}" not found.`);
        }
    });
}


/**
 * Initializes all Tabulator tables
 */
function initializeAllTabulatorTables() {
    // Define table configurations
    const tableConfigs = [
        {
            tableId: "trusted-characters-table",
            indexField: "CharacterID",
            data: TrustedCharacters,
            columns: [
                { title: "Character Name", field: "CharacterName", headerSort: true },
                { title: "Corporation", field: "CorporationName", headerSort: true },
                {
                    title: "Remove",
                    formatter: "buttonCross",
                    width: 100,
                    hozAlign: "center",
                    headerSort: false,
                    cellClick: function (e, cell) {
                        const rowData = cell.getRow().getData();
                        const characterName = rowData.CharacterName;
                        const characterID = rowData.CharacterID;
                        console.log(`Removing trusted character with ID: ${characterID}, Name: ${characterName}`);

                        // Use SweetAlert2 for Confirmation
                        Swal.fire({
                            title: `Remove Character?`,
                            text: `Do you want to stop trusting "${characterName}"?`,
                            icon: 'warning',
                            showCancelButton: true,
                            confirmButtonColor: '#d33',
                            cancelButtonColor: '#3085d6',
                            confirmButtonText: 'Yes',
                            cancelButtonText: 'No'
                        }).then((result) => {
                            if (result.isConfirmed) {
                                removeEntity('trusted', 'character', characterID.toString()); // Convert to string
                            }
                        });
                    }
                }
            ]
        },
        {
            tableId: "trusted-corporations-table",
            indexField: "CorporationID",
            data: TrustedCorporations,
            columns: [
                { title: "Corporation Name", field: "CorporationName", headerSort: true },
                { title: "Alliance Name", field: "AllianceName", headerSort: true },
                {
                    title: "Remove",
                    formatter: "buttonCross",
                    width: 100,
                    hozAlign: "center",
                    headerSort: false,
                    cellClick: function (e, cell) {
                        const rowData = cell.getRow().getData();
                        const corporationID = rowData.CorporationID;
                        const corporationName = rowData.CorporationName;
                        console.log(`Removing trusted corporation with ID: ${corporationID}, Name: ${corporationName}`);

                        // Use SweetAlert2 for Confirmation
                        Swal.fire({
                            title: `Remove Corporation?`,
                            text: `Do you want to stop trusting "${corporationName}"?`,
                            icon: 'warning',
                            showCancelButton: true,
                            confirmButtonColor: '#d33',
                            cancelButtonColor: '#3085d6',
                            confirmButtonText: 'Yes',
                            cancelButtonText: 'No'
                        }).then((result) => {
                            if (result.isConfirmed) {
                                removeEntity('trusted', 'corporation', corporationID.toString()); // Convert to string
                            }
                        });
                    }
                }
            ]
        },
        {
            tableId: "untrusted-characters-table",
            indexField: "CharacterID",
            data: UntrustedCharacters,
            columns: [
                { title: "Character Name", field: "CharacterName", headerSort: true },
                { title: "Corporation", field: "CorporationName", headerSort: true },
                {
                    title: "Remove",
                    formatter: "buttonCross",
                    width: 100,
                    hozAlign: "center",
                    headerSort: false,
                    cellClick: function (e, cell) {
                        const rowData = cell.getRow().getData();
                        const characterName = rowData.CharacterName;
                        const characterID = rowData.CharacterID;
                        console.log(`Removing untrusted character with ID: ${characterID}, Name: ${characterName}`);

                        // Use SweetAlert2 for Confirmation
                        Swal.fire({
                            title: `Remove Character?`,
                            text: `Has everyone updated their standings for "${characterName}"?`,
                            icon: 'warning',
                            showCancelButton: true,
                            confirmButtonColor: '#d33',
                            cancelButtonColor: '#3085d6',
                            confirmButtonText: 'Yes',
                            cancelButtonText: 'No'
                        }).then((result) => {
                            if (result.isConfirmed) {
                                removeEntity('untrusted', 'character', characterID.toString()); // Convert to string
                            }
                        });
                    }
                }
            ]
        },
        {
            tableId: "untrusted-corporations-table",
            indexField: "CorporationID",
            data: UntrustedCorporations,
            columns: [
                { title: "Corporation Name", field: "CorporationName", headerSort: true },
                { title: "Alliance Name", field: "AllianceName", headerSort: true },
                {
                    title: "Remove",
                    formatter: "buttonCross",
                    width: 100,
                    hozAlign: "center",
                    headerSort: false,
                    cellClick: function (e, cell) {
                        const rowData = cell.getRow().getData();
                        const corporationID = rowData.CorporationID;
                        const corporationName = rowData.CorporationName;
                        console.log(`Removing untrusted corporation with ID: ${corporationID}, Name: ${corporationName}`);

                        // Use SweetAlert2 for Confirmation
                        Swal.fire({
                            title: `Remove Corporation?`,
                            text: `Has everyone updated their standings for "${corporationName}"?`,
                            icon: 'warning',
                            showCancelButton: true,
                            confirmButtonColor: '#d33',
                            cancelButtonColor: '#3085d6',
                            confirmButtonText: 'Yes',
                            cancelButtonText: 'No'
                        }).then((result) => {
                            if (result.isConfirmed) {
                                removeEntity('untrusted', 'corporation', corporationID.toString()); // Convert to string
                            }
                        });
                    }
                }
            ]
        }
    ];

    // Initialize each table using the generic function
    tableConfigs.forEach(config => {
        initializeTabulatorTable(config.tableId, config.indexField, config.data, config.columns);
    });

    // Setup Mutation Observers for all tables
    const allTableIds = tableConfigs.map(config => config.tableId);
    setupMutationObservers(allTableIds);
}

/**
 * Removes an entity based on trustStatus and entityType using a single identifier.
 * @param {string} trustStatus - 'trusted' or 'untrusted'.
 * @param {string} entityType - 'character' or 'corporation'.
 * @param {string|number} identifier - Character/Corporation ID or Name.
 */
async function removeEntity(trustStatus, entityType, identifier) {
    console.log("Removing entity:", trustStatus, entityType, identifier);

    const serverEndpoint = getServerEndpoint(trustStatus, entityType, 'remove');
    if (!serverEndpoint) {
        console.error("Invalid trustStatus or entityType provided to removeEntity.");
        toastr.error("An unexpected error occurred.");
        return;
    }

    // Ensure identifier is always a string
    const identifierStr = String(identifier);

    // Prepare payload
    const payload = { identifier: identifierStr };

    try {
        showLoading();
        const response = await fetch(serverEndpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || "Failed to remove entity.");
        }

        console.log(`Entity removed successfully: ${identifierStr}`);


        const tableId = `${trustStatus}-${entityType}s-table`;
        removeRowFromTable(tableId, identifierStr);

        // Update local data and the table
        removeUpdateLocalData(trustStatus, entityType, identifierStr);

        if (trustStatus === 'trusted') {
            addEntity("untrusted", entityType, identifier);
        }

    } catch (error) {
        console.error(`Error removing ${entityType}: ${error}`);
        toastr.error(`Failed to remove ${entityType}. ${error.message}`);
    } finally {
        hideLoading();
    }
}

/**
 * Removes entity data from the appropriate local list.
 * @param {string} trustStatus - 'trusted' or 'untrusted'
 * @param {string} entityType - 'character' or 'corporation'
 * @param {string|number} identifier - The identifier of the entity to remove
 */
function removeUpdateLocalData(trustStatus, entityType, identifier) {
    const identifierNum = Number(identifier); // Ensures compatibility if identifier is a string

    // Determine which list to modify directly
    if (trustStatus === 'trusted' && entityType === 'character') {
        TrustedCharacters = TrustedCharacters.filter(entity => entity.CharacterID !== identifierNum);
    } else if (trustStatus === 'trusted' && entityType === 'corporation') {
        TrustedCorporations = TrustedCorporations.filter(entity => entity.CorporationID !== identifierNum);
    } else if (trustStatus === 'untrusted' && entityType === 'character') {
        UntrustedCharacters = UntrustedCharacters.filter(entity => entity.CharacterID !== identifierNum);
    } else if (trustStatus === 'untrusted' && entityType === 'corporation') {
        UntrustedCorporations = UntrustedCorporations.filter(entity => entity.CorporationID !== identifierNum);
    } else {
        console.warn(`Unknown trustStatus (${trustStatus}) or entityType (${entityType})`);
        return;
    }

    console.log(`Removed ${entityType} with ID ${identifier} from ${trustStatus} list.`);
}


/**
 * Removes a row from the specified table by identifier.
 * @param {string} tableId - The ID of the table.
 * @param {string|number} identifier - The identifier of the row to remove.
 */
function removeRowFromTable(tableId, identifier) {
    const targetTable = tables[tableId];
    if (targetTable) {
        const numericIdentifier = Number(identifier);
        const entityType = tableId.includes('character') ? 'Character' : 'Corporation';
        const trustStatus = tableId.split('-')[0];
        const data = targetTable.getData().find(row => row.CharacterID === numericIdentifier || row.CorporationID === numericIdentifier);
        targetTable.deleteRow(numericIdentifier)
            .then(() => {
                resizeTabulatorTable(tableId);
                if (targetTable.getData().length === 0) {
                    toggleSectionDisplay(tableId, false);
                }
                toastr.success(`Removed ${data[`${entityType}Name`]} from ${capitalize(trustStatus)} list.`);
            })
            .catch(error => {
                console.error(`Error deleting row from ${tableId}:`, error);
                const entityType = tableId.includes('character') ? 'Character' : 'Corporation';
                const trustStatus = tableId.split('-')[0];
                toastr.error(`Error updating ${data[`${entityType}Name`]} in ${capitalize(trustStatus)} ${entityType} table.`);
            });
    } else {
        console.warn(`Table with ID "${tableId}" not found.`);
    }
}

/**
 * Capitalizes the first letter of a string
 * @param {string} str - The string to capitalize.
 * @returns {string} - The capitalized string.
 */
function capitalize(str) {
    if (typeof str !== 'string') return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
}

/**
 * Checks if a string is numeric.
 * @param {string} str - The string to check.
 * @returns {boolean} - True if numeric, else false.
 */
function isNumeric(str) {
    return /^\d+$/.test(str);
}


/**
 * Setup Mutation Observers for tables
 * @param {Array} tableIds - Array of table IDs to observe
 */
function setupMutationObservers(tableIds) {
    tableIds.forEach(tableId => {
        const tableElement = document.getElementById(tableId);
        if (!tableElement) {
            console.warn(`Table element with ID "${tableId}" not found for MutationObserver.`);
            return;
        }

        new MutationObserver(function(mutations) {
            mutations.forEach(function(mutation) {
                if (mutation.attributeName === "style") {
                    const tableDisplay = window.getComputedStyle(tableElement).display;
                    console.log(`Style mutation detected for ${tableId}: ${tableDisplay}`);

                    const isTrustedTable = tableId.startsWith('trusted-');
                    const isUntrustedTable = tableId.startsWith('untrusted-');

                    if (isTrustedTable && !isShowingUntrusted) {
                        // Show trusted tables
                        toggleSectionDisplay(tableId, true);
                    } else if (isUntrustedTable && isShowingUntrusted) {
                        // Show untrusted tables
                        toggleSectionDisplay(tableId, true);
                    } else {
                        // Hide tables not currently in view
                        toggleSectionDisplay(tableId, false);
                    }
                }
            });
        }).observe(tableElement, { attributes: true });
    });
}

/**
 * Initializes a Tabulator table with given parameters
 * @param {string} tableId - The ID of the table container
 * @param {string} indexField - The field to use as row index
 * @param {Array} data - The initial data array
 * @param {Array} columns - The column definitions
 */
function initializeTabulatorTable(tableId, indexField, data, columns) {
    tables[tableId] = new Tabulator(`#${tableId}`, {
        index: indexField,
        data: data || [],
        layout: "fitColumns",
        responsiveLayout: "hide",
        placeholder: `No ${capitalize(tableId.split('-')[1].slice(0, -1))}s`,
        columns: columns,
        rowAdded: function () {
            resizeTabulatorTable(tableId);
        },
        rowDeleted: function () {
            resizeTabulatorTable(tableId);
        },
        tableBuilt: function () {
            // Ensure resize happens after the table is fully built
            resizeTabulatorTable(tableId);
        }
    });
}


/**
 * Toggle Button Event Listener
 * Switches between showing trusted and untrusted sections
 */
function setupToggleButton() {
    const toggleBtn = document.getElementById("toggle-contacts-btn");
    if (toggleBtn) {
        toggleBtn.addEventListener("click", function () {
            console.log("Toggle button clicked.");

            const trustedSections = [
                "trusted-characters-table",
                "trusted-corporations-table",
                "add-trusted-character-section",
                "add-trusted-corporation-section"
            ];

            const untrustedSections = [
                "untrusted-characters-table",
                "untrusted-corporations-table",
                "add-untrusted-character-section",
                "add-untrusted-corporation-section"
            ];

            // Determine the current state using the `isShowingUntrusted` flag
            const icon = this.querySelector("i");

            if (isShowingUntrusted) {
                console.log("Switching to trusted view...");

                // Show Trusted Tables and Forms, Hide Untrusted
                toggleMultipleSections(trustedSections, true);
                toggleMultipleSections(untrustedSections, false);

                icon.classList.remove("fa-toggle-off");
                icon.classList.add("fa-toggle-on");
                this.title = "Show Untrusted Contacts";

                isShowingUntrusted = false;

            } else {
                console.log("Switching to untrusted view...");

                // Show Untrusted Tables and Forms, Hide Trusted
                toggleMultipleSections(untrustedSections, true);
                toggleMultipleSections(trustedSections, false);

                icon.classList.remove("fa-toggle-on");
                icon.classList.add("fa-toggle-off");
                this.title = "Show Trusted Contacts";

                isShowingUntrusted = true;
            }

            // Resize all visible tables after toggling
            setTimeout(() => {
                const tableIds = [
                    "trusted-characters-table",
                    "trusted-corporations-table",
                    "untrusted-characters-table",
                    "untrusted-corporations-table"
                ];

                tableIds.forEach(tableId => {
                    const tableElement = document.getElementById(tableId);
                    if (tableElement && window.getComputedStyle(tableElement).display !== "none") {
                        resizeTabulatorTable(tableId);
                    }
                });
            }, 200); // Delay to allow the DOM to update
        });
    } else {
        console.error(`Toggle button with ID "toggle-contacts-btn" not found.`);
    }
}

/**
 * Initialize Everything After DOM is Loaded
 */
document.addEventListener('DOMContentLoaded', function () {
    // Set initial state of the view
    isShowingUntrusted = false;  // Ensure trusted tables visible and untrusted hidden.

    // Hide untrusted tables and sections immediately on page load
    const untrustedSections = [
        "untrusted-characters-table",
        "untrusted-corporations-table",
        "add-untrusted-character-section",
        "add-untrusted-corporation-section"
    ];
    toggleMultipleSections(untrustedSections, false);

    // Initialize all Tabulator tables
    initializeAllTabulatorTables();

    // Show trusted sections if they have data
    const trustedSections = [
        "trusted-characters-table",
        "trusted-corporations-table"
    ];
    trustedSections.forEach(tableId => {
        const hasData = tables[tableId].getData().length > 0;
        toggleSectionDisplay(tableId, hasData);
    });

    // Initialize character tiles
    initializeCharacterTiles();

    // Setup Toggle Button Event Listener
    setupToggleButton();

    // Initial resizing of tables on page load
    setTimeout(() => {
        const initialTables = [
            "trusted-characters-table",
            "trusted-corporations-table"
            // Untrusted tables are hidden on page load
        ];

        initialTables.forEach(tableId => {
            const tableElement = document.getElementById(tableId);
            if (tableElement && window.getComputedStyle(tableElement).display !== "none") {
                resizeTabulatorTable(tableId);
            }
        });

        // Untrusted tables and sections are already hidden earlier
    }, 200);

    // Setup all form event listeners
    setupFormEventListeners();
});
