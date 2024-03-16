

class State {
    constructor() {
        this._userid= "",
        this._extensionToken = "",
        this._allowTemporaryMemory= true,
        this._recordNavigation= false,
        this._automaticSearch= true,
        this._memorySize= 10,
        this._memory= []
    }

    get userId() {
        return this._userid;
    }
    set userId(id) {
        this._userid = id;
    }

    get extensionToken() {
        return this._extensionToken;
    }
    set extensionToken(token) {
        this._extensionToken = token;
    }

    get allowTemporaryMemory() {
        return this._allowTemporaryMemory;
    }
    set allowTemporaryMemory(remember) {
        this._allowTemporaryMemory = remember;
    }

    get recordNavigation() {
        return this._recordNavigation;
    }
    set recordNavigation(record) {
        this._recordNavigation = record;
    }

    get automaticSearch() {
        return this._automaticSearch;
    }
    set automaticSearch(search) {
        this._automaticSearch = search;
    }

    get memory() {
        if (this._allowTemporaryMemory) {
            return this._memory;
        }else{
            console.warn("Temporary memory is off, but you are trying to access it");
            return [];
        }
    }
    pushToMemory(page) {
        this._memory.push(page);
        if (this._memory.length > this._memorySize) {
            this._memory.shift();
        }
    }

    serialize() {
        return {
            userId: this._userid,
            extensionToken: this._extensionToken,            
            allowTemporaryMemory: this._allowTemporaryMemory,
            recordNavigation: this._recordNavigation,
            automaticSearch: this._automaticSearch,
            memorySize: this._memorySize,
            memory: this._memory
        };
    }
    static deserialize(obj) {
        const state = new State(
            obj.userId,
            obj.extensionToken,
            obj.allowTemporaryMemory,
            obj.recordNavigation,
            obj.automaticSearch,
            obj.memorySize,
            obj.memory
        );
        return state;
    }
};

const B = browser || chrome;
// get state from memory
var storedState = B.storage.local.get("lastState").then((res) => {
    if (!res.lastState) {
        B.storage.local.set({"lastState": (new State()).serialize()});
        return new State();
    }else{
        return State.deserialize(res.lastState);
    }
})


