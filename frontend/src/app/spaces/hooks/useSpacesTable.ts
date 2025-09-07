import { useState, useCallback } from 'react';
import { safeTimestampToDate } from '@/lib/types';

type SortKey = 'title' | 'stats.contentCount' | 'createdAt' | 'updatedAt';

export function useSpacesTable() {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editedField, setEditedField] = useState<string | null>(null);
  const [editedValue, setEditedValue] = useState<string | string[] | null>(null);
  const [sortConfig, setSortConfig] = useState<{ key: SortKey; direction: 'ascending' | 'descending' } | null>(null);

  const handleEdit = useCallback((spaceId: string, field: string, value: string | string[]) => {
    setEditingId(spaceId);
    setEditedField(field);
    setEditedValue(value);
  }, []);

  const handleSave = useCallback(async (_spaceId: string) => {
    if (editingId && editedField && editedValue !== null) {
      // Note: In a real implementation, this would call an API to update the space
      // For now, we're just clearing the editing state
      setEditingId(null);
      setEditedField(null);
      setEditedValue(null);
    }
  }, [editingId, editedField, editedValue]);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    if (editedField === 'keywords') {
      setEditedValue(e.target.value.split(',').map(k => k.trim()));
    } else {
      setEditedValue(e.target.value);
    }
  }, [editedField]);

  const requestSort = useCallback((key: SortKey) => {
    let direction: 'ascending' | 'descending' = 'ascending';
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    }
    setSortConfig({ key, direction });
  }, [sortConfig]);

  const getSortIcon = useCallback((key: SortKey) => {
    if (!sortConfig || sortConfig.key !== key) {
      return 'sortable';
    }
    return sortConfig.direction === 'ascending' ? 'asc' : 'desc';
  }, [sortConfig]);


  const getAccessColor = useCallback((_level: string) => {
    // TODO: Implement access level logic based on new data model
    return 'bg-green-100 text-green-800 border-green-200';
  }, []);

  const formatTimestamp = useCallback((timestamp: unknown) => {
    const date = safeTimestampToDate(timestamp);
    return date ? date.toLocaleDateString() : '—';
  }, []);

  return {
    editingId,
    editedField,
    editedValue,
    sortConfig,
    handleEdit,
    handleSave,
    handleChange,
    requestSort,
    getSortIcon,
    getAccessColor,
    formatTimestamp,
  };
}
