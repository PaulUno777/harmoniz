export namespace domain {
	
	export class ArtistSuggestion {
	    original: string;
	    suggested: string;
	    score: number;
	    reason: string;
	    confidence_level: string;
	
	    static createFrom(source: any = {}) {
	        return new ArtistSuggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.original = source["original"];
	        this.suggested = source["suggested"];
	        this.score = source["score"];
	        this.reason = source["reason"];
	        this.confidence_level = source["confidence_level"];
	    }
	}
	export class Track {
	    id: number;
	    path: string;
	    filename: string;
	    size: number;
	    mod_time: number;
	    added_at: number;
	    artist_raw: string;
	    artist_norm: string;
	    album_raw: string;
	    album_norm: string;
	    title: string;
	    year: number;
	    track_num: number;
	    bitrate: number;
	    hash_partial: string;
	    hash_full: string;
	    fingerprint: string;
	    is_deleted: boolean;
	    deleted_at?: number;
	    delete_reason: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new Track(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.filename = source["filename"];
	        this.size = source["size"];
	        this.mod_time = source["mod_time"];
	        this.added_at = source["added_at"];
	        this.artist_raw = source["artist_raw"];
	        this.artist_norm = source["artist_norm"];
	        this.album_raw = source["album_raw"];
	        this.album_norm = source["album_norm"];
	        this.title = source["title"];
	        this.year = source["year"];
	        this.track_num = source["track_num"];
	        this.bitrate = source["bitrate"];
	        this.hash_partial = source["hash_partial"];
	        this.hash_full = source["hash_full"];
	        this.fingerprint = source["fingerprint"];
	        this.is_deleted = source["is_deleted"];
	        this.deleted_at = source["deleted_at"];
	        this.delete_reason = source["delete_reason"];
	        this.status = source["status"];
	    }
	}
	export class DuplicateGroup {
	    tracks: Track[];
	    recommended_keep_id: number;
	
	    static createFrom(source: any = {}) {
	        return new DuplicateGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tracks = this.convertValues(source["tracks"], Track);
	        this.recommended_keep_id = source["recommended_keep_id"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TraceStep {
	    step: string;
	    input: string;
	    result: string;
	    confidence: number;
	    rejected?: boolean;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new TraceStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.step = source["step"];
	        this.input = source["input"];
	        this.result = source["result"];
	        this.confidence = source["confidence"];
	        this.rejected = source["rejected"];
	        this.reason = source["reason"];
	    }
	}
	export class FieldSuggestion {
	    value: string;
	    confidence: number;
	    source: string;
	    trace?: TraceStep[];
	
	    static createFrom(source: any = {}) {
	        return new FieldSuggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.confidence = source["confidence"];
	        this.source = source["source"];
	        this.trace = this.convertValues(source["trace"], TraceStep);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FilenamePreview {
	    track_id: number;
	    before: string;
	    after: string;
	
	    static createFrom(source: any = {}) {
	        return new FilenamePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.track_id = source["track_id"];
	        this.before = source["before"];
	        this.after = source["after"];
	    }
	}
	export class ListTracksResult {
	    Tracks: Track[];
	    Total: number;
	
	    static createFrom(source: any = {}) {
	        return new ListTracksResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Tracks = this.convertValues(source["Tracks"], Track);
	        this.Total = source["Total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OrganizerSuggestion {
	    track_id: number;
	    track: Track;
	    score: number;
	    fields: Record<string, FieldSuggestion>;
	    issues: string[];
	
	    static createFrom(source: any = {}) {
	        return new OrganizerSuggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.track_id = source["track_id"];
	        this.track = this.convertValues(source["track"], Track);
	        this.score = source["score"];
	        this.fields = this.convertValues(source["fields"], FieldSuggestion, true);
	        this.issues = source["issues"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace update {
	
	export class Result {
	    currentVersion: string;
	    latestVersion: string;
	    updateAvailable: boolean;
	    releasePageUrl: string;
	    downloadUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.updateAvailable = source["updateAvailable"];
	        this.releasePageUrl = source["releasePageUrl"];
	        this.downloadUrl = source["downloadUrl"];
	    }
	}

}

