'use client';

import { useState, ReactNode } from 'react';
import { HelpCircle } from 'lucide-react';

interface TooltipProps {
  content: string;
  children?: ReactNode;
  trigger?: 'icon' | 'text';
  position?: 'top' | 'bottom' | 'left' | 'right';
}

export default function Tooltip({ 
  content, 
  children, 
  trigger = 'icon',
  position = 'top' 
}: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false);

  const positionClasses = {
    top: 'bottom-full left-1/2 transform -translate-x-1/2 mb-2',
    bottom: 'top-full left-1/2 transform -translate-x-1/2 mt-2',
    left: 'right-full top-1/2 transform -translate-y-1/2 mr-2',
    right: 'left-full top-1/2 transform -translate-y-1/2 ml-2'
  };

  return (
    <span 
      className="relative inline-block"
      onMouseEnter={() => setIsVisible(true)}
      onMouseLeave={() => setIsVisible(false)}
      onFocus={() => setIsVisible(true)}
      onBlur={() => setIsVisible(false)}
    >
      {trigger === 'icon' ? (
        <button
          className="text-blue-600 hover:text-blue-800 focus:outline-none focus:ring-2 focus:ring-blue-300 rounded-full p-1"
          aria-label="Mostrar explicação"
          tabIndex={0}
        >
          <HelpCircle className="h-4 w-4" aria-hidden="true" />
        </button>
      ) : (
        <span 
          className="underline decoration-dotted cursor-help text-blue-600 hover:text-blue-800"
          tabIndex={0}
        >
          {children}
        </span>
      )}

      {isVisible && (
        <div
          className={`
            absolute z-50 px-3 py-2 text-sm text-white bg-gray-900 rounded-lg shadow-lg
            max-w-xs w-max ${positionClasses[position]}
          `}
          role="tooltip"
        >
          {content}
          <div 
            className={`
              absolute w-2 h-2 bg-gray-900 rotate-45
              ${position === 'top' ? 'top-full left-1/2 transform -translate-x-1/2 -mt-1' : ''}
              ${position === 'bottom' ? 'bottom-full left-1/2 transform -translate-x-1/2 -mb-1' : ''}
              ${position === 'left' ? 'left-full top-1/2 transform -translate-y-1/2 -ml-1' : ''}
              ${position === 'right' ? 'right-full top-1/2 transform -translate-y-1/2 -mr-1' : ''}
            `}
            aria-hidden="true"
          />
        </div>
      )}
    </span>
  );
}