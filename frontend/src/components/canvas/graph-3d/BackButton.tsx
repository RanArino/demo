'use client';

import React from 'react';

export function BackButton({ onBack, isVisible }: { onBack: () => void; isVisible: boolean }) {
  if (!isVisible) return null;

  return (
    <div className="absolute top-4 left-4 z-10">
      <button
        onClick={onBack}
        className="px-4 py-2 bg-blue-500 text-white rounded-lg shadow-lg hover:bg-blue-600 transition-colors"
      >
        ← 戻る
      </button>
    </div>
  );
} 