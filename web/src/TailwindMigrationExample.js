// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * This is a reference implementation showing how to migrate from Ant Design to Tailwind + shadcn/ui
 * 
 * Migration Pattern:
 * 1. Import shadcn/ui components instead of Ant Design
 * 2. Replace Ant Design components with shadcn/ui equivalents
 * 3. Use Tailwind utility classes for styling
 * 4. Maintain the same functionality and user experience
 * 
 * Example Migrations:
 * - antd Button -> shadcn Button
 * - antd Card -> shadcn Card  
 * - antd Table -> shadcn Table
 * - antd Input -> shadcn Input
 */

import React from "react";
import {Button} from "./components/ui/button";
import {Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter} from "./components/ui/card";
import {Input} from "./components/ui/input";
import {Table, TableHeader, TableBody, TableHead, TableRow, TableCell} from "./components/ui/table";

// This is a demonstration component showing the migration pattern
const TailwindMigrationExample = () => {
  const sampleData = [
    {id: 1, name: "Application 1", organization: "Org A", createdTime: "2024-01-01"},
    {id: 2, name: "Application 2", organization: "Org B", createdTime: "2024-01-02"},
    {id: 3, name: "Application 3", organization: "Org C", createdTime: "2024-01-03"},
  ];

  return (
    <div className="container mx-auto p-8">
      <h1 className="text-3xl font-bold mb-6">Tailwind + shadcn/ui Migration Example</h1>
      
      {/* Card Example - replaces antd Card */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Welcome to the New UI</CardTitle>
          <CardDescription>
            This demonstrates how Ant Design components are migrated to Tailwind + shadcn/ui
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {/* Input Example - replaces antd Input */}
            <div>
              <label className="block text-sm font-medium mb-2">Search Applications</label>
              <Input placeholder="Enter application name..." />
            </div>
            
            {/* Button Examples - replaces antd Button */}
            <div className="flex gap-2">
              <Button variant="primary">Primary Action</Button>
              <Button variant="outline">Secondary Action</Button>
              <Button variant="destructive">Delete</Button>
              <Button variant="ghost">Cancel</Button>
            </div>
          </div>
        </CardContent>
        <CardFooter>
          <p className="text-sm text-slate-500">Card footer content</p>
        </CardFooter>
      </Card>

      {/* Table Example - replaces antd Table */}
      <Card>
        <CardHeader>
          <CardTitle>Applications List</CardTitle>
          <CardDescription>Example of migrated table component</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Organization</TableHead>
                <TableHead>Created Time</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sampleData.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="font-medium">{item.name}</TableCell>
                  <TableCell>{item.organization}</TableCell>
                  <TableCell>{item.createdTime}</TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm">Edit</Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Migration Notes */}
      <Card className="mt-6 bg-blue-50 border-blue-200">
        <CardHeader>
          <CardTitle className="text-blue-900">Migration Notes</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-blue-800 space-y-2">
            <p><strong>Component Mapping:</strong></p>
            <ul className="list-disc list-inside space-y-1">
              <li>antd Button → shadcn/ui Button (with variants: primary, outline, ghost, destructive)</li>
              <li>antd Card → shadcn/ui Card (with CardHeader, CardTitle, CardContent, CardFooter)</li>
              <li>antd Table → shadcn/ui Table (with TableHeader, TableBody, TableRow, TableCell)</li>
              <li>antd Input → shadcn/ui Input</li>
              <li>antd Form → React Hook Form or native forms with Tailwind styling</li>
              <li>antd Modal → shadcn/ui Dialog</li>
              <li>antd Select → shadcn/ui Select</li>
            </ul>
            <p className="mt-4"><strong>Styling Approach:</strong></p>
            <ul className="list-disc list-inside space-y-1">
              <li>Use Tailwind utility classes for custom styling</li>
              <li>Leverage shadcn/ui's built-in variants</li>
              <li>Use cn() utility to merge classNames conditionally</li>
              <li>Maintain consistent spacing with Tailwind's spacing scale</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default TailwindMigrationExample;
