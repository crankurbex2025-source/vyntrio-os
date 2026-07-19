import type { ReactNode } from "react";

export type ApplianceDataTableColumn = {
  key: string;
  header: string;
  className?: string;
};

type ApplianceDataTableProps = {
  columns: ApplianceDataTableColumn[];
  rows: Array<Record<string, ReactNode>>;
  emptyMessage?: string;
  caption?: string;
};

export function ApplianceDataTable({
  columns,
  rows,
  emptyMessage = "No rows.",
  caption,
}: ApplianceDataTableProps) {
  if (rows.length === 0) {
    return <p className="vyn-ops-empty">{emptyMessage}</p>;
  }

  return (
    <div className="vyn-ops-table-wrap">
      <table className="vyn-ops-table">
        {caption ? <caption className="vyn-ops-sr-only">{caption}</caption> : null}
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column.key} scope="col" className={column.className}>
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, index) => (
            <tr key={index}>
              {columns.map((column) => (
                <td key={column.key} className={column.className}>
                  {row[column.key]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
