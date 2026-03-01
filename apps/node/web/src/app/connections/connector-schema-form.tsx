"use client";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type SchemaProperty = {
  type?: string;
  title?: string;
  description?: string;
  default?: unknown;
  enum?: string[];
  items?: { type?: string };
  "x-order"?: number;
  "x-secret"?: boolean;
  "x-multiline"?: boolean;
  "x-group"?: string;
  "x-visible-when"?: Record<string, string>;
};

type Spec = {
  type?: string;
  required?: string[];
  properties?: Record<string, SchemaProperty>;
};

type Props = {
  spec: Record<string, unknown>;
  values: Record<string, unknown>;
  onChange: (values: Record<string, unknown>) => void;
};

function humanize(key: string): string {
  return key
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

export function ConnectorSchemaForm({ spec, values, onChange }: Props) {
  const s = spec as Spec;
  const properties = s.properties ?? {};
  const requiredFields = new Set(s.required ?? []);

  const entries = Object.entries(properties).sort(
    ([, a], [, b]) => (a["x-order"] ?? 999) - (b["x-order"] ?? 999)
  );

  function setValue(key: string, value: unknown) {
    onChange({ ...values, [key]: value });
  }

  // Group fields by x-group
  const groups = new Map<string, [string, SchemaProperty][]>();
  const ungrouped: [string, SchemaProperty][] = [];

  for (const entry of entries) {
    const group = entry[1]["x-group"];
    if (group) {
      if (!groups.has(group)) {
        groups.set(group, []);
      }
      groups.get(group)!.push(entry);
    } else {
      ungrouped.push(entry);
    }
  }

  function isVisible(prop: SchemaProperty): boolean {
    const condition = prop["x-visible-when"];
    if (!condition) return true;
    return Object.entries(condition).every(
      ([field, expected]) => values[field] === expected
    );
  }

  function renderField(key: string, prop: SchemaProperty) {
    if (!isVisible(prop)) return null;

    const label = prop.title ?? humanize(key);
    const isRequired = requiredFields.has(key);
    const fieldId = `schema-${key}`;
    const currentValue = values[key];

    // Boolean → checkbox
    if (prop.type === "boolean") {
      return (
        <div key={key} className="flex items-center gap-2">
          <input
            id={fieldId}
            type="checkbox"
            checked={currentValue === true}
            onChange={(e) => setValue(key, e.target.checked)}
            className="h-4 w-4 rounded border-gray-300"
          />
          <Label htmlFor={fieldId}>{label}</Label>
          {prop.description ? (
            <span className="text-xs text-muted-foreground">{prop.description}</span>
          ) : null}
        </div>
      );
    }

    // Enum → select
    if (prop.enum) {
      return (
        <div key={key} className="space-y-2">
          <Label htmlFor={fieldId}>
            {label}
            {isRequired ? <span className="text-destructive"> *</span> : null}
          </Label>
          <select
            id={fieldId}
            value={(currentValue as string) ?? ""}
            onChange={(e) => setValue(key, e.target.value)}
            required={isRequired}
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option value="">Select...</option>
            {prop.enum.map((opt) => (
              <option key={opt} value={opt}>
                {opt}
              </option>
            ))}
          </select>
          {prop.description ? (
            <p className="text-xs text-muted-foreground">{prop.description}</p>
          ) : null}
        </div>
      );
    }

    // Multiline → textarea
    if (prop["x-multiline"]) {
      return (
        <div key={key} className="space-y-2">
          <Label htmlFor={fieldId}>
            {label}
            {isRequired ? <span className="text-destructive"> *</span> : null}
          </Label>
          <textarea
            id={fieldId}
            value={(currentValue as string) ?? ""}
            onChange={(e) => setValue(key, e.target.value)}
            required={isRequired}
            rows={4}
            className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
          {prop.description ? (
            <p className="text-xs text-muted-foreground">{prop.description}</p>
          ) : null}
        </div>
      );
    }

    // Number/integer → number input
    if (prop.type === "integer" || prop.type === "number") {
      return (
        <div key={key} className="space-y-2">
          <Label htmlFor={fieldId}>
            {label}
            {isRequired ? <span className="text-destructive"> *</span> : null}
          </Label>
          <Input
            id={fieldId}
            type="number"
            value={currentValue !== undefined && currentValue !== null ? String(currentValue) : ""}
            onChange={(e) => {
              const v = e.target.value;
              setValue(key, v === "" ? undefined : Number(v));
            }}
            required={isRequired}
          />
          {prop.description ? (
            <p className="text-xs text-muted-foreground">{prop.description}</p>
          ) : null}
        </div>
      );
    }

    // Array of strings → comma-separated input
    if (prop.type === "array" && prop.items?.type === "string") {
      const arr = Array.isArray(currentValue) ? currentValue : [];
      return (
        <div key={key} className="space-y-2">
          <Label htmlFor={fieldId}>
            {label}
            {isRequired ? <span className="text-destructive"> *</span> : null}
          </Label>
          <Input
            id={fieldId}
            value={arr.join(", ")}
            onChange={(e) => {
              const v = e.target.value;
              setValue(
                key,
                v.trim() === ""
                  ? []
                  : v.split(",").map((s) => s.trim())
              );
            }}
            placeholder="value1, value2, ..."
            required={isRequired}
          />
          {prop.description ? (
            <p className="text-xs text-muted-foreground">{prop.description}</p>
          ) : null}
        </div>
      );
    }

    // Default: string input (with x-secret support)
    return (
      <div key={key} className="space-y-2">
        <Label htmlFor={fieldId}>
          {label}
          {isRequired ? <span className="text-destructive"> *</span> : null}
        </Label>
        <Input
          id={fieldId}
          type={prop["x-secret"] ? "password" : "text"}
          value={(currentValue as string) ?? ""}
          onChange={(e) => setValue(key, e.target.value)}
          required={isRequired}
        />
        {prop.description ? (
          <p className="text-xs text-muted-foreground">{prop.description}</p>
        ) : null}
      </div>
    );
  }

  function renderGroup(name: string, fields: [string, SchemaProperty][]) {
    const visibleFields = fields.filter(([, prop]) => isVisible(prop));
    if (visibleFields.length === 0) return null;

    return (
      <div key={name} className="space-y-4 rounded-md border p-4">
        <h3 className="text-sm font-medium">{humanize(name)}</h3>
        <div className="grid gap-4 md:grid-cols-2">
          {visibleFields.map(([key, prop]) => renderField(key, prop))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2">
        {ungrouped.map(([key, prop]) => renderField(key, prop))}
      </div>
      {[...groups.entries()].map(([name, fields]) => renderGroup(name, fields))}
    </div>
  );
}
