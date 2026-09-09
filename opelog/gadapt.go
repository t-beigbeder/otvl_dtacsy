package opelog

import (
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

func gr2ser(gr *opeloggrpc.Rights) *Rights {
	return &Rights{Read: gr.Read, Write: gr.Write, Execute: gr.Execute}
}

func GrpcStoredEntry2StoredEntry(gse *opeloggrpc.StoredEntry) *StoredEntry {
	if gse == nil {
		return nil
	}
	children := make([]string, len(gse.Children))
	copy(children, gse.Children)
	return &StoredEntry{
		IsPresent:     gse.IsPresent,
		IsDir:         gse.IsDir,
		Size:          gse.Size,
		Mtime:         gse.Mtime,
		User:          gse.User,
		UserRights:    gr2ser(gse.UserRights),
		Group:         gse.Group,
		GroupRights:   gr2ser(gse.GroupRights),
		OtherRights:   gr2ser(gse.OtherRights),
		IsSymLink:     gse.IsSymLink,
		SymLinkTarget: gse.SymLinkTarget,
		Children:      children,
	}
}

func gses2ses(gses []*opeloggrpc.StoredEntry) []*StoredEntry {
	if gses == nil {
		return nil
	}
	ses := make([]*StoredEntry, len(gses))
	for i, gse := range gses {
		ses[i] = GrpcStoredEntry2StoredEntry(gse)
	}
	return ses
}

func GrpcEvent2Event(ge *opeloggrpc.Event) *Event {
	if ge == nil {
		return nil
	}
	return &Event{
		Kind:       EventCode(ge.Kind),
		Origin:     OriginCode(ge.Origin),
		TimeStamp:  ge.TimeStamp,
		StateIndex: ge.StateIndex,
		Checksums:  ge.Checksums,
		Error:      ge.Error,
	}
}

func gevs2evs(gevs []*opeloggrpc.Event) []*Event {
	if gevs == nil {
		return nil
	}
	evs := make([]*Event, len(gevs))
	for i, gev := range gevs {
		evs[i] = GrpcEvent2Event(gev)
	}
	return evs
}

func gvr2vr(gvr *opeloggrpc.Verification) *Verification {
	if gvr == nil {
		return nil
	}
	return &Verification{
		TimeStamp:    gvr.TimeStamp,
		WithChecksum: gvr.WithChecksum,
		NewStatus:    GrpcStoredEntry2StoredEntry(gvr.NewStatus),
		NewChecksums: gvr.NewChecksums,
	}
}

func gai2ai(gai *opeloggrpc.AggInfo) *AggInfo {
	if gai == nil {
		return nil
	}
	return &AggInfo{Number: gai.Number, Size: gai.Size}
}

func GrpcComputedStats2ComputedStats(gcs *opeloggrpc.ComputedStats) *ComputedStats {
	if gcs == nil {
		return nil
	}
	return &ComputedStats{
		TimeStamp:        gcs.TimeStamp,
		SourceListOrStat: gai2ai(gcs.SourceListOrStat),
		TargetListOrStat: gai2ai(gcs.TargetListOrStat),
		Read:             gai2ai(gcs.Read),
		Create:           gai2ai(gcs.Create),
		Update:           gai2ai(gcs.Update),
		Remove:           gai2ai(gcs.Remove),
		ModChange:        gai2ai(gcs.ModChange),
		Error:            gai2ai(gcs.Error),
	}
}

func gcss2css(gcss []*opeloggrpc.ComputedStats) []*ComputedStats {
	if gcss == nil {
		return nil
	}
	css := make([]*ComputedStats, len(gcss))
	for i, gcs := range gcss {
		css[i] = GrpcComputedStats2ComputedStats(gcs)
	}
	return css
}

func GrpcLogicalEntry2LogicalEntry(gle *opeloggrpc.LogicalEntry) *LogicalEntry {
	if gle == nil {
		return nil
	}
	dirupChildren := make([]string, len(gle.DirupChildren))
	copy(dirupChildren, gle.DirupChildren)
	return &LogicalEntry{
		SourceStates:  gses2ses(gle.SourceStates),
		SourceEvents:  gevs2evs(gle.SourceEvents),
		SourceVerif:   gvr2vr(gle.SourceVerif),
		TargetStates:  gses2ses(gle.TargetStates),
		DirupState:    GrpcStoredEntry2StoredEntry(gle.DirupState),
		DirupChildren: dirupChildren,
		TargetEvents:  gevs2evs(gle.TargetEvents),
		TargetVerif:   gvr2vr(gle.TargetVerif),
		StatsList:     gcss2css(gle.StatsList),
	}
}

func ser2gr(ser *Rights) *opeloggrpc.Rights {
	return &opeloggrpc.Rights{Read: ser.Read, Write: ser.Write, Execute: ser.Execute}
}

func StoredEntry2GrpcStoredEntry(se *StoredEntry) *opeloggrpc.StoredEntry {
	if se == nil {
		return nil
	}
	children := make([]string, len(se.Children))
	copy(children, se.Children)
	return &opeloggrpc.StoredEntry{
		IsPresent:     se.IsPresent,
		IsDir:         se.IsDir,
		Size:          se.Size,
		Mtime:         se.Mtime,
		User:          se.User,
		UserRights:    ser2gr(se.UserRights),
		Group:         se.Group,
		GroupRights:   ser2gr(se.GroupRights),
		OtherRights:   ser2gr(se.OtherRights),
		IsSymLink:     se.IsSymLink,
		SymLinkTarget: se.SymLinkTarget,
		Children:      children,
	}
}

func ses2gses(ses []*StoredEntry) []*opeloggrpc.StoredEntry {
	if ses == nil {
		return nil
	}
	gses := make([]*opeloggrpc.StoredEntry, len(ses))
	for i, se := range ses {
		gses[i] = StoredEntry2GrpcStoredEntry(se)
	}
	return gses
}

func Event2GrpcEvent(ev *Event) *opeloggrpc.Event {
	if ev == nil {
		return nil
	}
	return &opeloggrpc.Event{
		Kind:       opeloggrpc.EventCode(ev.Kind),
		Origin:     opeloggrpc.OriginCode(ev.Origin),
		TimeStamp:  ev.TimeStamp,
		StateIndex: ev.StateIndex,
		Checksums:  ev.Checksums,
		Error:      ev.Error,
	}
}

func evs2gevs(evs []*Event) []*opeloggrpc.Event {
	if evs == nil {
		return nil
	}
	gevs := make([]*opeloggrpc.Event, len(evs))
	for i, ev := range evs {
		gevs[i] = Event2GrpcEvent(ev)
	}
	return gevs
}

func vr2gvr(vr *Verification) *opeloggrpc.Verification {
	if vr == nil {
		return nil
	}
	return &opeloggrpc.Verification{
		TimeStamp:    vr.TimeStamp,
		WithChecksum: vr.WithChecksum,
		NewStatus:    StoredEntry2GrpcStoredEntry(vr.NewStatus),
		NewChecksums: vr.NewChecksums,
	}
}

func ai2gai(ai *AggInfo) *opeloggrpc.AggInfo {
	if ai == nil {
		return nil
	}
	return &opeloggrpc.AggInfo{Number: ai.Number, Size: ai.Size}
}

func ComputedStats2GrpcComputedStats(cs *ComputedStats) *opeloggrpc.ComputedStats {
	if cs == nil {
		return nil
	}
	return &opeloggrpc.ComputedStats{
		TimeStamp:        cs.TimeStamp,
		SourceListOrStat: ai2gai(cs.SourceListOrStat),
		TargetListOrStat: ai2gai(cs.TargetListOrStat),
		Read:             ai2gai(cs.Read),
		Create:           ai2gai(cs.Create),
		Update:           ai2gai(cs.Update),
		Remove:           ai2gai(cs.Remove),
		ModChange:        ai2gai(cs.ModChange),
		Error:            ai2gai(cs.Error),
	}
}

func css2gcss(css []*ComputedStats) []*opeloggrpc.ComputedStats {
	if css == nil {
		return nil
	}
	gcss := make([]*opeloggrpc.ComputedStats, len(css))
	for i, cs := range css {
		gcss[i] = ComputedStats2GrpcComputedStats(cs)
	}
	return gcss
}

func LogicalEntry2GrpcLogicalEntry(le *LogicalEntry) *opeloggrpc.LogicalEntry {
	if le == nil {
		return nil
	}
	dirupChildren := make([]string, len(le.DirupChildren))
	copy(dirupChildren, le.DirupChildren)
	return &opeloggrpc.LogicalEntry{
		SourceStates:  ses2gses(le.SourceStates),
		SourceEvents:  evs2gevs(le.SourceEvents),
		SourceVerif:   vr2gvr(le.SourceVerif),
		TargetStates:  ses2gses(le.TargetStates),
		DirupState:    StoredEntry2GrpcStoredEntry(le.DirupState),
		DirupChildren: dirupChildren,
		TargetEvents:  evs2gevs(le.TargetEvents),
		TargetVerif:   vr2gvr(le.TargetVerif),
		StatsList:     css2gcss(le.StatsList),
	}
}
