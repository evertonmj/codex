export interface Options { file?: string; encryptionKey?: string; ledger?: boolean; }
/** Synchronous embedded database. Always close the database to release its file lock. */
export class CodexDatabase {
  constructor(options?: Options);
  set(key: string, value: unknown): boolean;
  get<T = unknown>(key: string): T;
  delete(key: string): boolean;
  has(key: string): boolean;
  keys(): string[];
  clear(): boolean;
  close(): void;
}
