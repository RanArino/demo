import React from 'react';
import { Globe, Lock, Eye } from 'lucide-react';

interface AccessIconProps {
  level: string;
}

export function AccessIcon({ level }: AccessIconProps) {
  // TODO: Implement access level logic based on new data model
  // For now, defaulting to public access
  switch (level.toLowerCase()) {
    case 'private':
      return <Lock className="h-4 w-4 text-red-600" />;
    case 'restricted':
      return <Eye className="h-4 w-4 text-yellow-600" />;
    case 'public':
    default:
      return <Globe className="h-4 w-4 text-green-600" />;
  }
}
