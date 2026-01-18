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

import React, {useState} from "react";
import {Button} from "./components/ui/button";
import {Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter} from "./components/ui/card";
import {Input} from "./components/ui/input";
import {Table, TableHeader, TableBody, TableHead, TableRow, TableCell} from "./components/ui/table";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger} from "./components/ui/dialog";
import {Alert, AlertDescription, AlertTitle} from "./components/ui/alert";
import {Badge} from "./components/ui/badge";
import {AlertCircle, CheckCircle, Info} from "lucide-react";

// This is a demonstration component showing the migration pattern
const TailwindMigrationExample = () => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  
  const sampleData = [
    {id: 1, name: "Application 1", organization: "Org A", createdTime: "2024-01-01", status: "active"},
    {id: 2, name: "Application 2", organization: "Org B", createdTime: "2024-01-02", status: "pending"},
    {id: 3, name: "Application 3", organization: "Org C", createdTime: "2024-01-03", status: "inactive"},
  ];

  return (
    <div className="container mx-auto p-8">
      <h1 className="text-3xl font-bold mb-6">Tailwind + shadcn/ui Migration Example</h1>
      
      {/* Alert Examples - replaces antd Alert */}
      <div className="space-y-4 mb-6">
        <Alert variant="success">
          <CheckCircle className="h-4 w-4" />
          <AlertTitle>Success</AlertTitle>
          <AlertDescription>
            Tailwind CSS has been successfully integrated with shadcn/ui components.
          </AlertDescription>
        </Alert>
        
        <Alert variant="warning">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Warning</AlertTitle>
          <AlertDescription>
            Both Ant Design and Tailwind are active during the migration. Test thoroughly.
          </AlertDescription>
        </Alert>
        
        <Alert variant="info">
          <Info className="h-4 w-4" />
          <AlertTitle>Information</AlertTitle>
          <AlertDescription>
            Check the migration guide for detailed instructions on component migration.
          </AlertDescription>
        </Alert>
      </div>
      
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
            <div className="flex gap-2 flex-wrap">
              <Button variant="primary">Primary Action</Button>
              <Button variant="outline">Secondary Action</Button>
              <Button variant="destructive">Delete</Button>
              <Button variant="ghost">Cancel</Button>
              
              {/* Dialog Example - replaces antd Modal */}
              <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogTrigger asChild>
                  <Button variant="outline">Open Modal</Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Confirm Action</DialogTitle>
                    <DialogDescription>
                      This is a modal dialog example. It replaces Ant Design Modal component.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="py-4">
                    <p className="text-sm text-gray-600">
                      Are you sure you want to proceed with this action?
                    </p>
                  </div>
                  <DialogFooter>
                    <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                      Cancel
                    </Button>
                    <Button variant="primary" onClick={() => setIsDialogOpen(false)}>
                      Confirm
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>
        </CardContent>
        <CardFooter>
          <p className="text-sm text-slate-500">Card footer content</p>
        </CardFooter>
      </Card>

      {/* Badge Examples - replaces antd Badge */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Badge Examples</CardTitle>
          <CardDescription>Various badge styles for status indicators</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 flex-wrap">
            <Badge>Default</Badge>
            <Badge variant="secondary">Secondary</Badge>
            <Badge variant="destructive">Error</Badge>
            <Badge variant="success">Active</Badge>
            <Badge variant="warning">Pending</Badge>
            <Badge variant="outline">Outline</Badge>
          </div>
        </CardContent>
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
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sampleData.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="font-medium">{item.name}</TableCell>
                  <TableCell>{item.organization}</TableCell>
                  <TableCell>{item.createdTime}</TableCell>
                  <TableCell>
                    <Badge 
                      variant={
                        item.status === "active" ? "success" : 
                        item.status === "pending" ? "warning" : 
                        "secondary"
                      }
                    >
                      {item.status}
                    </Badge>
                  </TableCell>
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
              <li>antd Modal → shadcn/ui Dialog (with DialogTrigger, DialogContent, DialogHeader, DialogFooter)</li>
              <li>antd Alert → shadcn/ui Alert (with AlertTitle, AlertDescription)</li>
              <li>antd Badge → shadcn/ui Badge (with variants: default, secondary, destructive, success, warning)</li>
              <li>antd Form → React Hook Form or native forms with Tailwind styling</li>
              <li>antd Select → shadcn/ui Select (Radix UI based)</li>
            </ul>
            <p className="mt-4"><strong>Styling Approach:</strong></p>
            <ul className="list-disc list-inside space-y-1">
              <li>Use Tailwind utility classes for custom styling</li>
              <li>Leverage shadcn/ui's built-in variants</li>
              <li>Use cn() utility to merge classNames conditionally</li>
              <li>Maintain consistent spacing with Tailwind's spacing scale</li>
              <li>Use lucide-react for icons (replaces @ant-design/icons)</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default TailwindMigrationExample;
