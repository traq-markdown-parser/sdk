// Generated from Rust preset exports. Do not edit.
export type ProcessorPreset = "traq.v1";
export const processors = {"traq":{"v1":"traq.v1"}} as const;
export type ProcessorOptions = { 
/**
 * Origin used to recognize traQ file/message URLs; empty leaves URLs as text.
 */
origin: string, };
export type ProcessOutput = { notificationText: string, references: References, };
export type References = { mentions: Array<string>, groupMentions: Array<string>, channelLinks: Array<string>, };
