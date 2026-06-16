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

}

