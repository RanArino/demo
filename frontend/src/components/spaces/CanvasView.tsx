
import React, { useState } from 'react';
import { Space } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { 
  ArrowLeft, 
  ArrowRight, 
  Maximize2, 
  Settings, 
  Calendar,
  FileText,
  Box,
  Eye
} from 'lucide-react';
import { EditSpaceForm } from './EditSpaceForm';
import Link from 'next/link';

interface CanvasViewProps {
  spaces: Space[];
}

export function CanvasView({ spaces }: CanvasViewProps) {
  const [currentIndex, setCurrentIndex] = useState(0);

  const nextSpace = () => {
    setCurrentIndex((prev) => (prev + 1) % spaces.length);
  };

  const prevSpace = () => {
    setCurrentIndex((prev) => (prev - 1 + spaces.length) % spaces.length);
  };

  const currentSpace = spaces[currentIndex];

  if (spaces.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-96 text-center">
        <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
          <Box className="h-8 w-8 text-gray-400" />
        </div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">No spaces to display</h3>
        <p className="text-gray-500 max-w-sm">
          Create your first space to see it in the canvas view.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-medium text-gray-700">Canvas View</h2>
          <span className="text-sm text-gray-500">({spaces.length} spaces)</span>
        </div>
        
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500">
            {currentIndex + 1} of {spaces.length}
          </span>
        </div>
      </div>

      {/* Main Canvas Area */}
      <div className="relative bg-gradient-to-br from-blue-50 to-indigo-100 rounded-2xl p-8 min-h-[500px] overflow-hidden">
        {/* Background pattern */}
        <div className="absolute inset-0 opacity-5">
          <div className="absolute top-10 left-10 w-32 h-32 bg-blue-400 rounded-full blur-3xl"></div>
          <div className="absolute bottom-10 right-10 w-40 h-40 bg-purple-400 rounded-full blur-3xl"></div>
          <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-48 h-48 bg-pink-400 rounded-full blur-3xl"></div>
        </div>

        {/* Navigation Controls */}
        <div className="absolute left-4 top-1/2 transform -translate-y-1/2 z-10">
          <Button
            onClick={prevSpace}
            variant="secondary"
            size="icon"
            className="bg-white/80 hover:bg-white shadow-lg backdrop-blur-sm"
            disabled={spaces.length <= 1}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </div>
        
        <div className="absolute right-4 top-1/2 transform -translate-y-1/2 z-10">
          <Button
            onClick={nextSpace}
            variant="secondary"
            size="icon"
            className="bg-white/80 hover:bg-white shadow-lg backdrop-blur-sm"
            disabled={spaces.length <= 1}
          >
            <ArrowRight className="h-4 w-4" />
          </Button>
        </div>

        {/* Center Card */}
        <div className="flex items-center justify-center h-full">
          <div className="relative group">
            {/* Main Space Card */}
            <div className="bg-white rounded-2xl shadow-2xl p-8 max-w-md w-full transform transition-all duration-300 hover:scale-105">
              {/* Settings Button */}
              <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity">
                <Dialog>
                  <DialogTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <Settings className="h-4 w-4" />
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Edit Space</DialogTitle>
                    </DialogHeader>
                    <EditSpaceForm space={currentSpace} />
                  </DialogContent>
                </Dialog>
              </div>

              {/* Space Icon and Title */}
              <div className="text-center mb-6">
                <div className="text-6xl mb-4">{currentSpace.icon}</div>
                <h3 className="text-2xl font-bold text-gray-900 mb-2">{currentSpace.title}</h3>
                <p className="text-gray-600 leading-relaxed">{currentSpace.description}</p>
              </div>

              {/* Mind Map Preview Area */}
              <div className="bg-gray-50 rounded-xl p-6 mb-6 min-h-[200px] flex items-center justify-center border-2 border-dashed border-gray-200">
                <div className="text-center">
                  <Eye className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                  <p className="text-sm text-gray-500 mb-2">Mind Map Preview</p>
                  <div className="flex flex-wrap gap-1 justify-center">
                    {currentSpace.keywords.slice(0, 4).map((keyword) => (
                      <Badge key={keyword} variant="outline" className="text-xs">
                        {keyword}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>

              {/* Metadata */}
              <div className="space-y-3 text-sm text-gray-600">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Calendar className="h-4 w-4" />
                    <span>Created</span>
                  </div>
                  <span>{new Date(currentSpace.created_at).toLocaleDateString()}</span>
                </div>
                
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <FileText className="h-4 w-4" />
                    <span>Documents</span>
                  </div>
                  <span>{currentSpace.document_count}</span>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex gap-3 mt-6">
                <Dialog>
                  <DialogTrigger asChild>
                    <Button variant="outline" className="flex-1">
                      <Maximize2 className="h-4 w-4 mr-2" />
                      Quick Preview
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="max-w-4xl">
                    <DialogHeader>
                      <DialogTitle>Quick Preview - {currentSpace.title}</DialogTitle>
                    </DialogHeader>
                    <div className="p-6">
                      <div className="bg-gray-50 rounded-xl p-8 min-h-[400px] flex items-center justify-center border-2 border-dashed border-gray-200">
                        <div className="text-center">
                          <div className="text-8xl mb-4">{currentSpace.icon}</div>
                          <h3 className="text-3xl font-bold mb-4">{currentSpace.title}</h3>
                          <p className="text-gray-600 mb-6 max-w-md">{currentSpace.description}</p>
                          <div className="flex flex-wrap gap-2 justify-center mb-6">
                            {currentSpace.keywords.map((keyword) => (
                              <Badge key={keyword} variant="outline">
                                {keyword}
                              </Badge>
                            ))}
                          </div>
                          <Link href={`/spaces/${currentSpace.id}`}>
                            <Button className="bg-blue-600 hover:bg-blue-700">
                              Open Canvas
                            </Button>
                          </Link>
                        </div>
                      </div>
                    </div>
                  </DialogContent>
                </Dialog>
                
                <Link href={`/spaces/${currentSpace.id}`} className="flex-1">
                  <Button className="w-full bg-blue-600 hover:bg-blue-700">
                    Open Canvas
                  </Button>
                </Link>
              </div>
            </div>

            {/* Side Cards (Depth Effect) */}
            {spaces.length > 1 && (
              <>
                {/* Left side card */}
                <div className="absolute top-8 -left-12 bg-white rounded-xl shadow-lg p-4 w-48 opacity-60 transform rotate-6 scale-90 pointer-events-none">
                  <div className="text-center">
                    <div className="text-2xl mb-2">
                      {spaces[(currentIndex - 1 + spaces.length) % spaces.length]?.icon}
                    </div>
                    <h4 className="font-medium text-sm truncate">
                      {spaces[(currentIndex - 1 + spaces.length) % spaces.length]?.title}
                    </h4>
                  </div>
                </div>

                {/* Right side card */}
                <div className="absolute top-8 -right-12 bg-white rounded-xl shadow-lg p-4 w-48 opacity-60 transform -rotate-6 scale-90 pointer-events-none">
                  <div className="text-center">
                    <div className="text-2xl mb-2">
                      {spaces[(currentIndex + 1) % spaces.length]?.icon}
                    </div>
                    <h4 className="font-medium text-sm truncate">
                      {spaces[(currentIndex + 1) % spaces.length]?.title}
                    </h4>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>

        {/* Mini-map (placeholder) */}
        <div className="absolute bottom-4 left-4 bg-white/80 backdrop-blur-sm rounded-lg p-3 shadow-lg">
          <div className="text-xs text-gray-600 mb-2">Mini-map</div>
          <div className="flex gap-1">
            {spaces.map((_, index) => (
              <div
                key={index}
                className={`w-2 h-2 rounded-full cursor-pointer transition-colors ${
                  index === currentIndex ? 'bg-blue-600' : 'bg-gray-300'
                }`}
                onClick={() => setCurrentIndex(index)}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
