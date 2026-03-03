"use client";

import {
  Background,
  BaseEdge,
  getSmoothStepPath,
  Handle,
  Position,
  ReactFlow,
  type Edge,
  type EdgeProps,
  type Node,
  type NodeProps,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { Server, Terminal, Webhook, Zap } from "lucide-react";
import type { ReactNode } from "react";

function AnimatedEdge({
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  markerEnd,
}: EdgeProps) {
  const [edgePath] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });
  const guideLineColor = "var(--foreground)";
  const dottedLineColor = "var(--primary)";
  const particleColor = "var(--primary)";
  const streamOffsets = ["0s", "-0.5s", "-1s", "-1.5s"] as const;
  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          stroke: guideLineColor,
          strokeWidth: 2.5,
          opacity: 0.35,
        }}
      />
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          stroke: dottedLineColor,
          strokeWidth: 2.5,
          opacity: 0.95,
          strokeDasharray: "1 10",
          strokeLinecap: "round",
        }}
      />
      {streamOffsets.map((begin) => (
        <circle
          key={begin}
          r="4"
          fill={particleColor}
          opacity={0.95}
          style={{ filter: "drop-shadow(0 0 6px var(--primary))" }}
        >
          <animateMotion
            dur="1s"
            begin={begin}
            repeatCount="indefinite"
            path={edgePath}
          />
        </circle>
      ))}
    </>
  );
}

const edgeTypes = { animated: AnimatedEdge };

const NODE_WIDTH = 160;
const HORIZONTAL_GAP = 120;
const BRANCH_OFFSET_Y = 85;

interface FlowNodeData extends Record<string, unknown> {
  label: string;
  description: string;
  icon: ReactNode;
}

function FlowStepNode({ data }: NodeProps<Node<FlowNodeData>>) {
  return (
    <div className="px-5 py-4 rounded-xl border-2 border-border bg-card shadow-sm min-w-[160px] text-center">
      <Handle
        type="target"
        position={Position.Left}
        style={{
          opacity: 0,
          pointerEvents: "none",
          width: 0,
          height: 0,
          border: 0,
          background: "transparent",
        }}
        isConnectable={false}
      />
      <div className="flex justify-center mb-2 text-primary [&_svg]:size-6">
        {data.icon}
      </div>
      <div className="font-semibold text-sm">{data.label}</div>
      <div className="text-xs text-muted-foreground mt-0.5">
        {data.description}
      </div>
      <Handle
        type="source"
        position={Position.Right}
        style={{
          opacity: 0,
          pointerEvents: "none",
          width: 0,
          height: 0,
          border: 0,
          background: "transparent",
        }}
        isConnectable={false}
      />
    </div>
  );
}

const nodeTypes = { step: FlowStepNode };

const initialNodes: Node<FlowNodeData>[] = [
  {
    id: "sources",
    type: "step",
    position: { x: 0, y: 0 },
    data: {
      label: "Stripe, GitHub, …",
      description: "Webhook senders",
      icon: <Webhook className="size-6" />,
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
  },
  {
    id: "ingest",
    type: "step",
    position: { x: NODE_WIDTH + HORIZONTAL_GAP, y: 0 },
    data: {
      label: "Ingest",
      description: "Our endpoint",
      icon: <Server className="size-6" />,
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
  },
  {
    id: "relay",
    type: "step",
    position: { x: 2 * (NODE_WIDTH + HORIZONTAL_GAP), y: 0 },
    data: {
      label: "Relay",
      description: "Real-time delivery",
      icon: <Zap className="size-6" />,
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
  },
  {
    id: "richard-cli",
    type: "step",
    position: { x: 3 * (NODE_WIDTH + HORIZONTAL_GAP), y: -BRANCH_OFFSET_Y },
    data: {
      label: "Richard's machine",
      description: "CLI listening",
      icon: <Terminal className="size-6" />,
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
  },
  {
    id: "gilfoyle-cli",
    type: "step",
    position: { x: 3 * (NODE_WIDTH + HORIZONTAL_GAP), y: BRANCH_OFFSET_Y },
    data: {
      label: "Gilfoyle's machine",
      description: "CLI listening",
      icon: <Terminal className="size-6" />,
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
  },
];

const initialEdges: Edge[] = [
  {
    id: "e-sources-ingest",
    source: "sources",
    target: "ingest",
    type: "animated",
  },
  { id: "e-ingest-relay", source: "ingest", target: "relay", type: "animated" },
  {
    id: "e-relay-richard",
    source: "relay",
    target: "richard-cli",
    type: "animated",
  },
  {
    id: "e-relay-gilfoyle",
    source: "relay",
    target: "gilfoyle-cli",
    type: "animated",
  },
];

export function WebhookFlow() {
  return (
    <div className="w-full h-[360px] rounded-xl border border-border bg-muted/30">
      <ReactFlow
        nodes={initialNodes}
        edges={initialEdges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.4}
        maxZoom={1.2}
        zoomOnScroll={false}
        zoomOnPinch={false}
        zoomOnDoubleClick={false}
        panOnDrag={false}
        panOnScroll={false}
        panOnScrollSpeed={0}
        nodesDraggable={false}
        nodesConnectable={false}
        edgesFocusable={false}
        nodesFocusable={false}
        elementsSelectable={false}
        proOptions={{ hideAttribution: true }}
      >
        <Background gap={16} size={1} className="bg-muted/20" />
      </ReactFlow>
    </div>
  );
}
