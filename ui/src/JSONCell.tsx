import { Row } from '@tanstack/react-table';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { useState } from "react";
import { TaskEventInfo } from './types';

export const JsonCell = ({row}: { row: Row<TaskEventInfo>}) => {
    const [isOpen, setIsOpen] = useState(false);
    const data = JSON.stringify(row.getValue("data"), null, 2);
  
    return (
      <div className="w-full max-w-xs">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center justify-between w-full p-2 text-left bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
        >
          <span className="font-medium">View Data</span>
          {isOpen ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
        </button>
        {isOpen && (
          <pre className="mt-2 p-4 bg-gray-50 rounded-md overflow-x-auto text-sm">
            {data}
          </pre>
        )}
      </div>
    );
  };
  
export default JsonCell;  