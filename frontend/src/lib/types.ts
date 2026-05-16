export interface SkillItem {
  name: string;
  type: string;
  status: string;
  disabled: boolean;
  invocation: string;
  description: string;
  descriptionCN: string;
  triggers: string[];
  target: string;
}

export interface SkillDetailItem extends SkillItem {
  errors: string[];
}

export interface PluginItem {
  name: string;
  version: string;
  skills: SkillItem[];
  installPath: string;
  disabled: boolean;
}

export interface MemoryFileItem {
  file: string;
  name: string;
  type: string;
  description: string;
  content?: string;
}

export interface MemoryStats {
  total: number;
  byType: Record<string, number>;
  byProject: Record<string, number>;
  orphanCount: number;
  oldest: string;
  newest: string;
}

export interface MCPItem {
  name: string;
  project: string;
  command: string;
  args: string[];
  status: string;
  disabled: boolean;
}

export interface ClaudeMDItem {
  path: string;
  level: string;
  size: number;
  references: string[];
}

export interface PortabilityResult {
  issues: PortabilityIssue[];
  critical: number;
  warning: number;
  info: number;
}

export interface PortabilityIssue {
  severity: string;
  domain: string;
  message: string;
  fix: string;
}

export interface SecretItem {
  pattern: string;
  line: number;
  match: string;
  filePath: string;
}

export interface AppSettings {
  claudeDir: string;
  homeDir: string;
  autoStart: boolean;
  startMinimized: boolean;
}

export interface DashboardSummary {
  skillsCount: number;
  memoryCount: number;
  mcpServers: number;
  errorCount: number;
  warningCount: number;
}

export interface IssueItem {
  severity: string;
  domain: string;
  message: string;
  detail: string;
  fix: string;
}
