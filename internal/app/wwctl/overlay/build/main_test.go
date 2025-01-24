package build

import (
	"bytes"
	"path"
	"testing"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Benchmark_Overlay_Build(b *testing.B) {
	env := testenv.NewBenchmark(b)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf",
		`nodeprofiles:
  default:
    comment: This profile is automatically included for each node
    image name: rockylinux-9
    ipxe template: default
    runtime overlay:
      - hosts
      - ssh.authorized_keys
      - syncuser
    system overlay:
      - wwinit
      - wwclient
      - fstab
      - hostname
      - ssh.host_keys
      - issue
      - resolv
      - udev.netname
      - systemd.netname
      - ifcfg
      - NetworkManager
      - debian.interfaces
      - wicked
      - ignition
    kernel:
      args: quiet crashkernel=no
    ipmi:
      template: ipmitool.tmpl
    init: /sbin/init
    root: initramfs
nodes:
  tn1:
    profiles:
      - default
  tn2:
    profiles:
      - default
  tn3:
    profiles:
      - default
  tn4:
    profiles:
      - default
  tn5:
    profiles:
      - default
  tn6:
    profiles:
      - default
  tn7:
    profiles:
      - default
  tn8:
    profiles:
      - default
  tn9:
    profiles:
      - default
  tn10:
    profiles:
      - default
  tn11:
    profiles:
      - default
  tn12:
    profiles:
      - default
  tn13:
    profiles:
      - default
  tn14:
    profiles:
      - default
  tn15:
    profiles:
      - default
  tn16:
    profiles:
      - default
  tn17:
    profiles:
      - default
  tn18:
    profiles:
      - default
  tn19:
    profiles:
      - default
  tn20:
    profiles:
      - default
  tn21:
    profiles:
      - default
  tn22:
    profiles:
      - default
  tn23:
    profiles:
      - default
  tn24:
    profiles:
      - default
  tn25:
    profiles:
      - default
  tn26:
    profiles:
      - default
  tn27:
    profiles:
      - default
  tn28:
    profiles:
      - default
  tn29:
    profiles:
      - default
  tn30:
    profiles:
      - default
  tn31:
    profiles:
      - default
  tn32:
    profiles:
      - default
  tn33:
    profiles:
      - default
  tn34:
    profiles:
      - default
  tn35:
    profiles:
      - default
  tn36:
    profiles:
      - default
  tn37:
    profiles:
      - default
  tn38:
    profiles:
      - default
  tn39:
    profiles:
      - default
  tn40:
    profiles:
      - default
  tn41:
    profiles:
      - default
  tn42:
    profiles:
      - default
  tn43:
    profiles:
      - default
  tn44:
    profiles:
      - default
  tn45:
    profiles:
      - default
  tn46:
    profiles:
      - default
  tn47:
    profiles:
      - default
  tn48:
    profiles:
      - default
  tn49:
    profiles:
      - default
  tn50:
    profiles:
      - default
  tn51:
    profiles:
      - default
  tn52:
    profiles:
      - default
  tn53:
    profiles:
      - default
  tn54:
    profiles:
      - default
  tn55:
    profiles:
      - default
  tn56:
    profiles:
      - default
  tn57:
    profiles:
      - default
  tn58:
    profiles:
      - default
  tn59:
    profiles:
      - default
  tn60:
    profiles:
      - default
  tn61:
    profiles:
      - default
  tn62:
    profiles:
      - default
  tn63:
    profiles:
      - default
  tn64:
    profiles:
      - default
  tn65:
    profiles:
      - default
  tn66:
    profiles:
      - default
  tn67:
    profiles:
      - default
  tn68:
    profiles:
      - default
  tn69:
    profiles:
      - default
  tn70:
    profiles:
      - default
  tn71:
    profiles:
      - default
  tn72:
    profiles:
      - default
  tn73:
    profiles:
      - default
  tn74:
    profiles:
      - default
  tn75:
    profiles:
      - default
  tn76:
    profiles:
      - default
  tn77:
    profiles:
      - default
  tn78:
    profiles:
      - default
  tn79:
    profiles:
      - default
  tn80:
    profiles:
      - default
  tn81:
    profiles:
      - default
  tn82:
    profiles:
      - default
  tn83:
    profiles:
      - default
  tn84:
    profiles:
      - default
  tn85:
    profiles:
      - default
  tn86:
    profiles:
      - default
  tn87:
    profiles:
      - default
  tn88:
    profiles:
      - default
  tn89:
    profiles:
      - default
  tn90:
    profiles:
      - default
  tn91:
    profiles:
      - default
  tn92:
    profiles:
      - default
  tn93:
    profiles:
      - default
  tn94:
    profiles:
      - default
  tn95:
    profiles:
      - default
  tn96:
    profiles:
      - default
  tn97:
    profiles:
      - default
  tn98:
    profiles:
      - default
  tn99:
    profiles:
      - default
  tn100:
    profiles:
      - default
  tn101:
    profiles:
      - default
  tn102:
    profiles:
      - default
  tn103:
    profiles:
      - default
  tn104:
    profiles:
      - default
  tn105:
    profiles:
      - default
  tn106:
    profiles:
      - default
  tn107:
    profiles:
      - default
  tn108:
    profiles:
      - default
  tn109:
    profiles:
      - default
  tn110:
    profiles:
      - default
  tn111:
    profiles:
      - default
  tn112:
    profiles:
      - default
  tn113:
    profiles:
      - default
  tn114:
    profiles:
      - default
  tn115:
    profiles:
      - default
  tn116:
    profiles:
      - default
  tn117:
    profiles:
      - default
  tn118:
    profiles:
      - default
  tn119:
    profiles:
      - default
  tn120:
    profiles:
      - default
  tn121:
    profiles:
      - default
  tn122:
    profiles:
      - default
  tn123:
    profiles:
      - default
  tn124:
    profiles:
      - default
  tn125:
    profiles:
      - default
  tn126:
    profiles:
      - default
  tn127:
    profiles:
      - default
  tn128:
    profiles:
      - default
  tn129:
    profiles:
      - default
  tn130:
    profiles:
      - default
  tn131:
    profiles:
      - default
  tn132:
    profiles:
      - default
  tn133:
    profiles:
      - default
  tn134:
    profiles:
      - default
  tn135:
    profiles:
      - default
  tn136:
    profiles:
      - default
  tn137:
    profiles:
      - default
  tn138:
    profiles:
      - default
  tn139:
    profiles:
      - default
  tn140:
    profiles:
      - default
  tn141:
    profiles:
      - default
  tn142:
    profiles:
      - default
  tn143:
    profiles:
      - default
  tn144:
    profiles:
      - default
  tn145:
    profiles:
      - default
  tn146:
    profiles:
      - default
  tn147:
    profiles:
      - default
  tn148:
    profiles:
      - default
  tn149:
    profiles:
      - default
  tn150:
    profiles:
      - default
  tn151:
    profiles:
      - default
  tn152:
    profiles:
      - default
  tn153:
    profiles:
      - default
  tn154:
    profiles:
      - default
  tn155:
    profiles:
      - default
  tn156:
    profiles:
      - default
  tn157:
    profiles:
      - default
  tn158:
    profiles:
      - default
  tn159:
    profiles:
      - default
  tn160:
    profiles:
      - default
  tn161:
    profiles:
      - default
  tn162:
    profiles:
      - default
  tn163:
    profiles:
      - default
  tn164:
    profiles:
      - default
  tn165:
    profiles:
      - default
  tn166:
    profiles:
      - default
  tn167:
    profiles:
      - default
  tn168:
    profiles:
      - default
  tn169:
    profiles:
      - default
  tn170:
    profiles:
      - default
  tn171:
    profiles:
      - default
  tn172:
    profiles:
      - default
  tn173:
    profiles:
      - default
  tn174:
    profiles:
      - default
  tn175:
    profiles:
      - default
  tn176:
    profiles:
      - default
  tn177:
    profiles:
      - default
  tn178:
    profiles:
      - default
  tn179:
    profiles:
      - default
  tn180:
    profiles:
      - default
  tn181:
    profiles:
      - default
  tn182:
    profiles:
      - default
  tn183:
    profiles:
      - default
  tn184:
    profiles:
      - default
  tn185:
    profiles:
      - default
  tn186:
    profiles:
      - default
  tn187:
    profiles:
      - default
  tn188:
    profiles:
      - default
  tn189:
    profiles:
      - default
  tn190:
    profiles:
      - default
  tn191:
    profiles:
      - default
  tn192:
    profiles:
      - default
  tn193:
    profiles:
      - default
  tn194:
    profiles:
      - default
  tn195:
    profiles:
      - default
  tn196:
    profiles:
      - default
  tn197:
    profiles:
      - default
  tn198:
    profiles:
      - default
  tn199:
    profiles:
      - default
  tn200:
    profiles:
      - default
  tn201:
    profiles:
      - default
  tn202:
    profiles:
      - default
  tn203:
    profiles:
      - default
  tn204:
    profiles:
      - default
  tn205:
    profiles:
      - default
  tn206:
    profiles:
      - default
  tn207:
    profiles:
      - default
  tn208:
    profiles:
      - default
  tn209:
    profiles:
      - default
  tn210:
    profiles:
      - default
  tn211:
    profiles:
      - default
  tn212:
    profiles:
      - default
  tn213:
    profiles:
      - default
  tn214:
    profiles:
      - default
  tn215:
    profiles:
      - default
  tn216:
    profiles:
      - default
  tn217:
    profiles:
      - default
  tn218:
    profiles:
      - default
  tn219:
    profiles:
      - default
  tn220:
    profiles:
      - default
  tn221:
    profiles:
      - default
  tn222:
    profiles:
      - default
  tn223:
    profiles:
      - default
  tn224:
    profiles:
      - default
  tn225:
    profiles:
      - default
  tn226:
    profiles:
      - default
  tn227:
    profiles:
      - default
  tn228:
    profiles:
      - default
  tn229:
    profiles:
      - default
  tn230:
    profiles:
      - default
  tn231:
    profiles:
      - default
  tn232:
    profiles:
      - default
  tn233:
    profiles:
      - default
  tn234:
    profiles:
      - default
  tn235:
    profiles:
      - default
  tn236:
    profiles:
      - default
  tn237:
    profiles:
      - default
  tn238:
    profiles:
      - default
  tn239:
    profiles:
      - default
  tn240:
    profiles:
      - default
  tn241:
    profiles:
      - default
  tn242:
    profiles:
      - default
  tn243:
    profiles:
      - default
  tn244:
    profiles:
      - default
  tn245:
    profiles:
      - default
  tn246:
    profiles:
      - default
  tn247:
    profiles:
      - default
  tn248:
    profiles:
      - default
  tn249:
    profiles:
      - default
  tn250:
    profiles:
      - default
  tn251:
    profiles:
      - default
  tn252:
    profiles:
      - default
  tn253:
    profiles:
      - default
  tn254:
    profiles:
      - default
  tn255:
    profiles:
      - default
  tn256:
    profiles:
      - default
  tn257:
    profiles:
      - default
  tn258:
    profiles:
      - default
  tn259:
    profiles:
      - default
  tn260:
    profiles:
      - default
  tn261:
    profiles:
      - default
  tn262:
    profiles:
      - default
  tn263:
    profiles:
      - default
  tn264:
    profiles:
      - default
  tn265:
    profiles:
      - default
  tn266:
    profiles:
      - default
  tn267:
    profiles:
      - default
  tn268:
    profiles:
      - default
  tn269:
    profiles:
      - default
  tn270:
    profiles:
      - default
  tn271:
    profiles:
      - default
  tn272:
    profiles:
      - default
  tn273:
    profiles:
      - default
  tn274:
    profiles:
      - default
  tn275:
    profiles:
      - default
  tn276:
    profiles:
      - default
  tn277:
    profiles:
      - default
  tn278:
    profiles:
      - default
  tn279:
    profiles:
      - default
  tn280:
    profiles:
      - default
  tn281:
    profiles:
      - default
  tn282:
    profiles:
      - default
  tn283:
    profiles:
      - default
  tn284:
    profiles:
      - default
  tn285:
    profiles:
      - default
  tn286:
    profiles:
      - default
  tn287:
    profiles:
      - default
  tn288:
    profiles:
      - default
  tn289:
    profiles:
      - default
  tn290:
    profiles:
      - default
  tn291:
    profiles:
      - default
  tn292:
    profiles:
      - default
  tn293:
    profiles:
      - default
  tn294:
    profiles:
      - default
  tn295:
    profiles:
      - default
  tn296:
    profiles:
      - default
  tn297:
    profiles:
      - default
  tn298:
    profiles:
      - default
  tn299:
    profiles:
      - default
  tn300:
    profiles:
      - default
  tn301:
    profiles:
      - default
  tn302:
    profiles:
      - default
  tn303:
    profiles:
      - default
  tn304:
    profiles:
      - default
  tn305:
    profiles:
      - default
  tn306:
    profiles:
      - default
  tn307:
    profiles:
      - default
  tn308:
    profiles:
      - default
  tn309:
    profiles:
      - default
  tn310:
    profiles:
      - default
  tn311:
    profiles:
      - default
  tn312:
    profiles:
      - default
  tn313:
    profiles:
      - default
  tn314:
    profiles:
      - default
  tn315:
    profiles:
      - default
  tn316:
    profiles:
      - default
  tn317:
    profiles:
      - default
  tn318:
    profiles:
      - default
  tn319:
    profiles:
      - default
  tn320:
    profiles:
      - default
  tn321:
    profiles:
      - default
  tn322:
    profiles:
      - default
  tn323:
    profiles:
      - default
  tn324:
    profiles:
      - default
  tn325:
    profiles:
      - default
  tn326:
    profiles:
      - default
  tn327:
    profiles:
      - default
  tn328:
    profiles:
      - default
  tn329:
    profiles:
      - default
  tn330:
    profiles:
      - default
  tn331:
    profiles:
      - default
  tn332:
    profiles:
      - default
  tn333:
    profiles:
      - default
  tn334:
    profiles:
      - default
  tn335:
    profiles:
      - default
  tn336:
    profiles:
      - default
  tn337:
    profiles:
      - default
  tn338:
    profiles:
      - default
  tn339:
    profiles:
      - default
  tn340:
    profiles:
      - default
  tn341:
    profiles:
      - default
  tn342:
    profiles:
      - default
  tn343:
    profiles:
      - default
  tn344:
    profiles:
      - default
  tn345:
    profiles:
      - default
  tn346:
    profiles:
      - default
  tn347:
    profiles:
      - default
  tn348:
    profiles:
      - default
  tn349:
    profiles:
      - default
  tn350:
    profiles:
      - default
  tn351:
    profiles:
      - default
  tn352:
    profiles:
      - default
  tn353:
    profiles:
      - default
  tn354:
    profiles:
      - default
  tn355:
    profiles:
      - default
  tn356:
    profiles:
      - default
  tn357:
    profiles:
      - default
  tn358:
    profiles:
      - default
  tn359:
    profiles:
      - default
  tn360:
    profiles:
      - default
  tn361:
    profiles:
      - default
  tn362:
    profiles:
      - default
  tn363:
    profiles:
      - default
  tn364:
    profiles:
      - default
  tn365:
    profiles:
      - default
  tn366:
    profiles:
      - default
  tn367:
    profiles:
      - default
  tn368:
    profiles:
      - default
  tn369:
    profiles:
      - default
  tn370:
    profiles:
      - default
  tn371:
    profiles:
      - default
  tn372:
    profiles:
      - default
  tn373:
    profiles:
      - default
  tn374:
    profiles:
      - default
  tn375:
    profiles:
      - default
  tn376:
    profiles:
      - default
  tn377:
    profiles:
      - default
  tn378:
    profiles:
      - default
  tn379:
    profiles:
      - default
  tn380:
    profiles:
      - default
  tn381:
    profiles:
      - default
  tn382:
    profiles:
      - default
  tn383:
    profiles:
      - default
  tn384:
    profiles:
      - default
  tn385:
    profiles:
      - default
  tn386:
    profiles:
      - default
  tn387:
    profiles:
      - default
  tn388:
    profiles:
      - default
  tn389:
    profiles:
      - default
  tn390:
    profiles:
      - default
  tn391:
    profiles:
      - default
  tn392:
    profiles:
      - default
  tn393:
    profiles:
      - default
  tn394:
    profiles:
      - default
  tn395:
    profiles:
      - default
  tn396:
    profiles:
      - default
  tn397:
    profiles:
      - default
  tn398:
    profiles:
      - default
  tn399:
    profiles:
      - default
  tn400:
    profiles:
      - default
  tn401:
    profiles:
      - default
  tn402:
    profiles:
      - default
  tn403:
    profiles:
      - default
  tn404:
    profiles:
      - default
  tn405:
    profiles:
      - default
  tn406:
    profiles:
      - default
  tn407:
    profiles:
      - default
  tn408:
    profiles:
      - default
  tn409:
    profiles:
      - default
  tn410:
    profiles:
      - default
  tn411:
    profiles:
      - default
  tn412:
    profiles:
      - default
  tn413:
    profiles:
      - default
  tn414:
    profiles:
      - default
  tn415:
    profiles:
      - default
  tn416:
    profiles:
      - default
  tn417:
    profiles:
      - default
  tn418:
    profiles:
      - default
  tn419:
    profiles:
      - default
  tn420:
    profiles:
      - default
  tn421:
    profiles:
      - default
  tn422:
    profiles:
      - default
  tn423:
    profiles:
      - default
  tn424:
    profiles:
      - default
  tn425:
    profiles:
      - default
  tn426:
    profiles:
      - default
  tn427:
    profiles:
      - default
  tn428:
    profiles:
      - default
  tn429:
    profiles:
      - default
  tn430:
    profiles:
      - default
  tn431:
    profiles:
      - default
  tn432:
    profiles:
      - default
  tn433:
    profiles:
      - default
  tn434:
    profiles:
      - default
  tn435:
    profiles:
      - default
  tn436:
    profiles:
      - default
  tn437:
    profiles:
      - default
  tn438:
    profiles:
      - default
  tn439:
    profiles:
      - default
  tn440:
    profiles:
      - default
  tn441:
    profiles:
      - default
  tn442:
    profiles:
      - default
  tn443:
    profiles:
      - default
  tn444:
    profiles:
      - default
  tn445:
    profiles:
      - default
  tn446:
    profiles:
      - default
  tn447:
    profiles:
      - default
  tn448:
    profiles:
      - default
  tn449:
    profiles:
      - default
  tn450:
    profiles:
      - default
  tn451:
    profiles:
      - default
  tn452:
    profiles:
      - default
  tn453:
    profiles:
      - default
  tn454:
    profiles:
      - default
  tn455:
    profiles:
      - default
  tn456:
    profiles:
      - default
  tn457:
    profiles:
      - default
  tn458:
    profiles:
      - default
  tn459:
    profiles:
      - default
  tn460:
    profiles:
      - default
  tn461:
    profiles:
      - default
  tn462:
    profiles:
      - default
  tn463:
    profiles:
      - default
  tn464:
    profiles:
      - default
  tn465:
    profiles:
      - default
  tn466:
    profiles:
      - default
  tn467:
    profiles:
      - default
  tn468:
    profiles:
      - default
  tn469:
    profiles:
      - default
  tn470:
    profiles:
      - default
  tn471:
    profiles:
      - default
  tn472:
    profiles:
      - default
  tn473:
    profiles:
      - default
  tn474:
    profiles:
      - default
  tn475:
    profiles:
      - default
  tn476:
    profiles:
      - default
  tn477:
    profiles:
      - default
  tn478:
    profiles:
      - default
  tn479:
    profiles:
      - default
  tn480:
    profiles:
      - default
  tn481:
    profiles:
      - default
  tn482:
    profiles:
      - default
  tn483:
    profiles:
      - default
  tn484:
    profiles:
      - default
  tn485:
    profiles:
      - default
  tn486:
    profiles:
      - default
  tn487:
    profiles:
      - default
  tn488:
    profiles:
      - default
  tn489:
    profiles:
      - default
  tn490:
    profiles:
      - default
  tn491:
    profiles:
      - default
  tn492:
    profiles:
      - default
  tn493:
    profiles:
      - default
  tn494:
    profiles:
      - default
  tn495:
    profiles:
      - default
  tn496:
    profiles:
      - default
  tn497:
    profiles:
      - default
  tn498:
    profiles:
      - default
  tn499:
    profiles:
      - default
  tn500:
    profiles:
      - default
  tn501:
    profiles:
      - default
  tn502:
    profiles:
      - default
  tn503:
    profiles:
      - default
  tn504:
    profiles:
      - default
  tn505:
    profiles:
      - default
  tn506:
    profiles:
      - default
  tn507:
    profiles:
      - default
  tn508:
    profiles:
      - default
  tn509:
    profiles:
      - default
  tn510:
    profiles:
      - default
  tn511:
    profiles:
      - default
  tn512:
    profiles:
      - default
  tn513:
    profiles:
      - default
  tn514:
    profiles:
      - default
  tn515:
    profiles:
      - default
  tn516:
    profiles:
      - default
  tn517:
    profiles:
      - default
  tn518:
    profiles:
      - default
  tn519:
    profiles:
      - default
  tn520:
    profiles:
      - default
  tn521:
    profiles:
      - default
  tn522:
    profiles:
      - default
  tn523:
    profiles:
      - default
  tn524:
    profiles:
      - default
  tn525:
    profiles:
      - default
  tn526:
    profiles:
      - default
  tn527:
    profiles:
      - default
  tn528:
    profiles:
      - default
  tn529:
    profiles:
      - default
  tn530:
    profiles:
      - default
  tn531:
    profiles:
      - default
  tn532:
    profiles:
      - default
  tn533:
    profiles:
      - default
  tn534:
    profiles:
      - default
  tn535:
    profiles:
      - default
  tn536:
    profiles:
      - default
  tn537:
    profiles:
      - default
  tn538:
    profiles:
      - default
  tn539:
    profiles:
      - default
  tn540:
    profiles:
      - default
  tn541:
    profiles:
      - default
  tn542:
    profiles:
      - default
  tn543:
    profiles:
      - default
  tn544:
    profiles:
      - default
  tn545:
    profiles:
      - default
  tn546:
    profiles:
      - default
  tn547:
    profiles:
      - default
  tn548:
    profiles:
      - default
  tn549:
    profiles:
      - default
  tn550:
    profiles:
      - default
  tn551:
    profiles:
      - default
  tn552:
    profiles:
      - default
  tn553:
    profiles:
      - default
  tn554:
    profiles:
      - default
  tn555:
    profiles:
      - default
  tn556:
    profiles:
      - default
  tn557:
    profiles:
      - default
  tn558:
    profiles:
      - default
  tn559:
    profiles:
      - default
  tn560:
    profiles:
      - default
  tn561:
    profiles:
      - default
  tn562:
    profiles:
      - default
  tn563:
    profiles:
      - default
  tn564:
    profiles:
      - default
  tn565:
    profiles:
      - default
  tn566:
    profiles:
      - default
  tn567:
    profiles:
      - default
  tn568:
    profiles:
      - default
  tn569:
    profiles:
      - default
  tn570:
    profiles:
      - default
  tn571:
    profiles:
      - default
  tn572:
    profiles:
      - default
  tn573:
    profiles:
      - default
  tn574:
    profiles:
      - default
  tn575:
    profiles:
      - default
  tn576:
    profiles:
      - default
  tn577:
    profiles:
      - default
  tn578:
    profiles:
      - default
  tn579:
    profiles:
      - default
  tn580:
    profiles:
      - default
  tn581:
    profiles:
      - default
  tn582:
    profiles:
      - default
  tn583:
    profiles:
      - default
  tn584:
    profiles:
      - default
  tn585:
    profiles:
      - default
  tn586:
    profiles:
      - default
  tn587:
    profiles:
      - default
  tn588:
    profiles:
      - default
  tn589:
    profiles:
      - default
  tn590:
    profiles:
      - default
  tn591:
    profiles:
      - default
  tn592:
    profiles:
      - default
  tn593:
    profiles:
      - default
  tn594:
    profiles:
      - default
  tn595:
    profiles:
      - default
  tn596:
    profiles:
      - default
  tn597:
    profiles:
      - default
  tn598:
    profiles:
      - default
  tn599:
    profiles:
      - default
  tn600:
    profiles:
      - default
  tn601:
    profiles:
      - default
  tn602:
    profiles:
      - default
  tn603:
    profiles:
      - default
  tn604:
    profiles:
      - default
  tn605:
    profiles:
      - default
  tn606:
    profiles:
      - default
  tn607:
    profiles:
      - default
  tn608:
    profiles:
      - default
  tn609:
    profiles:
      - default
  tn610:
    profiles:
      - default
  tn611:
    profiles:
      - default
  tn612:
    profiles:
      - default
  tn613:
    profiles:
      - default
  tn614:
    profiles:
      - default
  tn615:
    profiles:
      - default
  tn616:
    profiles:
      - default
  tn617:
    profiles:
      - default
  tn618:
    profiles:
      - default
  tn619:
    profiles:
      - default
  tn620:
    profiles:
      - default
  tn621:
    profiles:
      - default
  tn622:
    profiles:
      - default
  tn623:
    profiles:
      - default
  tn624:
    profiles:
      - default
  tn625:
    profiles:
      - default
  tn626:
    profiles:
      - default
  tn627:
    profiles:
      - default
  tn628:
    profiles:
      - default
  tn629:
    profiles:
      - default
  tn630:
    profiles:
      - default
  tn631:
    profiles:
      - default
  tn632:
    profiles:
      - default
  tn633:
    profiles:
      - default
  tn634:
    profiles:
      - default
  tn635:
    profiles:
      - default
  tn636:
    profiles:
      - default
  tn637:
    profiles:
      - default
  tn638:
    profiles:
      - default
  tn639:
    profiles:
      - default
  tn640:
    profiles:
      - default
  tn641:
    profiles:
      - default
  tn642:
    profiles:
      - default
  tn643:
    profiles:
      - default
  tn644:
    profiles:
      - default
  tn645:
    profiles:
      - default
  tn646:
    profiles:
      - default
  tn647:
    profiles:
      - default
  tn648:
    profiles:
      - default
  tn649:
    profiles:
      - default
  tn650:
    profiles:
      - default
  tn651:
    profiles:
      - default
  tn652:
    profiles:
      - default
  tn653:
    profiles:
      - default
  tn654:
    profiles:
      - default
  tn655:
    profiles:
      - default
  tn656:
    profiles:
      - default
  tn657:
    profiles:
      - default
  tn658:
    profiles:
      - default
  tn659:
    profiles:
      - default
  tn660:
    profiles:
      - default
  tn661:
    profiles:
      - default
  tn662:
    profiles:
      - default
  tn663:
    profiles:
      - default
  tn664:
    profiles:
      - default
  tn665:
    profiles:
      - default
  tn666:
    profiles:
      - default
  tn667:
    profiles:
      - default
  tn668:
    profiles:
      - default
  tn669:
    profiles:
      - default
  tn670:
    profiles:
      - default
  tn671:
    profiles:
      - default
  tn672:
    profiles:
      - default
  tn673:
    profiles:
      - default
  tn674:
    profiles:
      - default
  tn675:
    profiles:
      - default
  tn676:
    profiles:
      - default
  tn677:
    profiles:
      - default
  tn678:
    profiles:
      - default
  tn679:
    profiles:
      - default
  tn680:
    profiles:
      - default
  tn681:
    profiles:
      - default
  tn682:
    profiles:
      - default
  tn683:
    profiles:
      - default
  tn684:
    profiles:
      - default
  tn685:
    profiles:
      - default
  tn686:
    profiles:
      - default
  tn687:
    profiles:
      - default
  tn688:
    profiles:
      - default
  tn689:
    profiles:
      - default
  tn690:
    profiles:
      - default
  tn691:
    profiles:
      - default
  tn692:
    profiles:
      - default
  tn693:
    profiles:
      - default
  tn694:
    profiles:
      - default
  tn695:
    profiles:
      - default
  tn696:
    profiles:
      - default
  tn697:
    profiles:
      - default
  tn698:
    profiles:
      - default
  tn699:
    profiles:
      - default
  tn700:
    profiles:
      - default
  tn701:
    profiles:
      - default
  tn702:
    profiles:
      - default
  tn703:
    profiles:
      - default
  tn704:
    profiles:
      - default
  tn705:
    profiles:
      - default
  tn706:
    profiles:
      - default
  tn707:
    profiles:
      - default
  tn708:
    profiles:
      - default
  tn709:
    profiles:
      - default
  tn710:
    profiles:
      - default
  tn711:
    profiles:
      - default
  tn712:
    profiles:
      - default
  tn713:
    profiles:
      - default
  tn714:
    profiles:
      - default
  tn715:
    profiles:
      - default
  tn716:
    profiles:
      - default
  tn717:
    profiles:
      - default
  tn718:
    profiles:
      - default
  tn719:
    profiles:
      - default
  tn720:
    profiles:
      - default
  tn721:
    profiles:
      - default
  tn722:
    profiles:
      - default
  tn723:
    profiles:
      - default
  tn724:
    profiles:
      - default
  tn725:
    profiles:
      - default
  tn726:
    profiles:
      - default
  tn727:
    profiles:
      - default
  tn728:
    profiles:
      - default
  tn729:
    profiles:
      - default
  tn730:
    profiles:
      - default
  tn731:
    profiles:
      - default
  tn732:
    profiles:
      - default
  tn733:
    profiles:
      - default
  tn734:
    profiles:
      - default
  tn735:
    profiles:
      - default
  tn736:
    profiles:
      - default
  tn737:
    profiles:
      - default
  tn738:
    profiles:
      - default
  tn739:
    profiles:
      - default
  tn740:
    profiles:
      - default
  tn741:
    profiles:
      - default
  tn742:
    profiles:
      - default
  tn743:
    profiles:
      - default
  tn744:
    profiles:
      - default
  tn745:
    profiles:
      - default
  tn746:
    profiles:
      - default
  tn747:
    profiles:
      - default
  tn748:
    profiles:
      - default
  tn749:
    profiles:
      - default
  tn750:
    profiles:
      - default
  tn751:
    profiles:
      - default
  tn752:
    profiles:
      - default
  tn753:
    profiles:
      - default
  tn754:
    profiles:
      - default
  tn755:
    profiles:
      - default
  tn756:
    profiles:
      - default
  tn757:
    profiles:
      - default
  tn758:
    profiles:
      - default
  tn759:
    profiles:
      - default
  tn760:
    profiles:
      - default
  tn761:
    profiles:
      - default
  tn762:
    profiles:
      - default
  tn763:
    profiles:
      - default
  tn764:
    profiles:
      - default
  tn765:
    profiles:
      - default
  tn766:
    profiles:
      - default
  tn767:
    profiles:
      - default
  tn768:
    profiles:
      - default
  tn769:
    profiles:
      - default
  tn770:
    profiles:
      - default
  tn771:
    profiles:
      - default
  tn772:
    profiles:
      - default
  tn773:
    profiles:
      - default
  tn774:
    profiles:
      - default
  tn775:
    profiles:
      - default
  tn776:
    profiles:
      - default
  tn777:
    profiles:
      - default
  tn778:
    profiles:
      - default
  tn779:
    profiles:
      - default
  tn780:
    profiles:
      - default
  tn781:
    profiles:
      - default
  tn782:
    profiles:
      - default
  tn783:
    profiles:
      - default
  tn784:
    profiles:
      - default
  tn785:
    profiles:
      - default
  tn786:
    profiles:
      - default
  tn787:
    profiles:
      - default
  tn788:
    profiles:
      - default
  tn789:
    profiles:
      - default
  tn790:
    profiles:
      - default
  tn791:
    profiles:
      - default
  tn792:
    profiles:
      - default
  tn793:
    profiles:
      - default
  tn794:
    profiles:
      - default
  tn795:
    profiles:
      - default
  tn796:
    profiles:
      - default
  tn797:
    profiles:
      - default
  tn798:
    profiles:
      - default
  tn799:
    profiles:
      - default
  tn800:
    profiles:
      - default
  tn801:
    profiles:
      - default
  tn802:
    profiles:
      - default
  tn803:
    profiles:
      - default
  tn804:
    profiles:
      - default
  tn805:
    profiles:
      - default
  tn806:
    profiles:
      - default
  tn807:
    profiles:
      - default
  tn808:
    profiles:
      - default
  tn809:
    profiles:
      - default
  tn810:
    profiles:
      - default
  tn811:
    profiles:
      - default
  tn812:
    profiles:
      - default
  tn813:
    profiles:
      - default
  tn814:
    profiles:
      - default
  tn815:
    profiles:
      - default
  tn816:
    profiles:
      - default
  tn817:
    profiles:
      - default
  tn818:
    profiles:
      - default
  tn819:
    profiles:
      - default
  tn820:
    profiles:
      - default
  tn821:
    profiles:
      - default
  tn822:
    profiles:
      - default
  tn823:
    profiles:
      - default
  tn824:
    profiles:
      - default
  tn825:
    profiles:
      - default
  tn826:
    profiles:
      - default
  tn827:
    profiles:
      - default
  tn828:
    profiles:
      - default
  tn829:
    profiles:
      - default
  tn830:
    profiles:
      - default
  tn831:
    profiles:
      - default
  tn832:
    profiles:
      - default
  tn833:
    profiles:
      - default
  tn834:
    profiles:
      - default
  tn835:
    profiles:
      - default
  tn836:
    profiles:
      - default
  tn837:
    profiles:
      - default
  tn838:
    profiles:
      - default
  tn839:
    profiles:
      - default
  tn840:
    profiles:
      - default
  tn841:
    profiles:
      - default
  tn842:
    profiles:
      - default
  tn843:
    profiles:
      - default
  tn844:
    profiles:
      - default
  tn845:
    profiles:
      - default
  tn846:
    profiles:
      - default
  tn847:
    profiles:
      - default
  tn848:
    profiles:
      - default
  tn849:
    profiles:
      - default
  tn850:
    profiles:
      - default
  tn851:
    profiles:
      - default
  tn852:
    profiles:
      - default
  tn853:
    profiles:
      - default
  tn854:
    profiles:
      - default
  tn855:
    profiles:
      - default
  tn856:
    profiles:
      - default
  tn857:
    profiles:
      - default
  tn858:
    profiles:
      - default
  tn859:
    profiles:
      - default
  tn860:
    profiles:
      - default
  tn861:
    profiles:
      - default
  tn862:
    profiles:
      - default
  tn863:
    profiles:
      - default
  tn864:
    profiles:
      - default
  tn865:
    profiles:
      - default
  tn866:
    profiles:
      - default
  tn867:
    profiles:
      - default
  tn868:
    profiles:
      - default
  tn869:
    profiles:
      - default
  tn870:
    profiles:
      - default
  tn871:
    profiles:
      - default
  tn872:
    profiles:
      - default
  tn873:
    profiles:
      - default
  tn874:
    profiles:
      - default
  tn875:
    profiles:
      - default
  tn876:
    profiles:
      - default
  tn877:
    profiles:
      - default
  tn878:
    profiles:
      - default
  tn879:
    profiles:
      - default
  tn880:
    profiles:
      - default
  tn881:
    profiles:
      - default
  tn882:
    profiles:
      - default
  tn883:
    profiles:
      - default
  tn884:
    profiles:
      - default
  tn885:
    profiles:
      - default
  tn886:
    profiles:
      - default
  tn887:
    profiles:
      - default
  tn888:
    profiles:
      - default
  tn889:
    profiles:
      - default
  tn890:
    profiles:
      - default
  tn891:
    profiles:
      - default
  tn892:
    profiles:
      - default
  tn893:
    profiles:
      - default
  tn894:
    profiles:
      - default
  tn895:
    profiles:
      - default
  tn896:
    profiles:
      - default
  tn897:
    profiles:
      - default
  tn898:
    profiles:
      - default
  tn899:
    profiles:
      - default
  tn900:
    profiles:
      - default
  tn901:
    profiles:
      - default
  tn902:
    profiles:
      - default
  tn903:
    profiles:
      - default
  tn904:
    profiles:
      - default
  tn905:
    profiles:
      - default
  tn906:
    profiles:
      - default
  tn907:
    profiles:
      - default
  tn908:
    profiles:
      - default
  tn909:
    profiles:
      - default
  tn910:
    profiles:
      - default
  tn911:
    profiles:
      - default
  tn912:
    profiles:
      - default
  tn913:
    profiles:
      - default
  tn914:
    profiles:
      - default
  tn915:
    profiles:
      - default
  tn916:
    profiles:
      - default
  tn917:
    profiles:
      - default
  tn918:
    profiles:
      - default
  tn919:
    profiles:
      - default
  tn920:
    profiles:
      - default
  tn921:
    profiles:
      - default
  tn922:
    profiles:
      - default
  tn923:
    profiles:
      - default
  tn924:
    profiles:
      - default
  tn925:
    profiles:
      - default
  tn926:
    profiles:
      - default
  tn927:
    profiles:
      - default
  tn928:
    profiles:
      - default
  tn929:
    profiles:
      - default
  tn930:
    profiles:
      - default
  tn931:
    profiles:
      - default
  tn932:
    profiles:
      - default
  tn933:
    profiles:
      - default
  tn934:
    profiles:
      - default
  tn935:
    profiles:
      - default
  tn936:
    profiles:
      - default
  tn937:
    profiles:
      - default
  tn938:
    profiles:
      - default
  tn939:
    profiles:
      - default
  tn940:
    profiles:
      - default
  tn941:
    profiles:
      - default
  tn942:
    profiles:
      - default
  tn943:
    profiles:
      - default
  tn944:
    profiles:
      - default
  tn945:
    profiles:
      - default
  tn946:
    profiles:
      - default
  tn947:
    profiles:
      - default
  tn948:
    profiles:
      - default
  tn949:
    profiles:
      - default
  tn950:
    profiles:
      - default
  tn951:
    profiles:
      - default
  tn952:
    profiles:
      - default
  tn953:
    profiles:
      - default
  tn954:
    profiles:
      - default
  tn955:
    profiles:
      - default
  tn956:
    profiles:
      - default
  tn957:
    profiles:
      - default
  tn958:
    profiles:
      - default
  tn959:
    profiles:
      - default
  tn960:
    profiles:
      - default
  tn961:
    profiles:
      - default
  tn962:
    profiles:
      - default
  tn963:
    profiles:
      - default
  tn964:
    profiles:
      - default
  tn965:
    profiles:
      - default
  tn966:
    profiles:
      - default
  tn967:
    profiles:
      - default
  tn968:
    profiles:
      - default
  tn969:
    profiles:
      - default
  tn970:
    profiles:
      - default
  tn971:
    profiles:
      - default
  tn972:
    profiles:
      - default
  tn973:
    profiles:
      - default
  tn974:
    profiles:
      - default
  tn975:
    profiles:
      - default
  tn976:
    profiles:
      - default
  tn977:
    profiles:
      - default
  tn978:
    profiles:
      - default
  tn979:
    profiles:
      - default
  tn980:
    profiles:
      - default
  tn981:
    profiles:
      - default
  tn982:
    profiles:
      - default
  tn983:
    profiles:
      - default
  tn984:
    profiles:
      - default
  tn985:
    profiles:
      - default
  tn986:
    profiles:
      - default
  tn987:
    profiles:
      - default
  tn988:
    profiles:
      - default
  tn989:
    profiles:
      - default
  tn990:
    profiles:
      - default
  tn991:
    profiles:
      - default
  tn992:
    profiles:
      - default
  tn993:
    profiles:
      - default
  tn994:
    profiles:
      - default
  tn995:
    profiles:
      - default
  tn996:
    profiles:
      - default
  tn997:
    profiles:
      - default
  tn998:
    profiles:
      - default
  tn999:
    profiles:
      - default`)

	runtimeOverlays := []string{"hosts", "ssh.authorized_keys", "syncuser"}
	systemOverlays := []string{"wwinit", "wwclient", "fstab", "hostname", "ssh.host_keys", "issue", "resolv", "udev.netname", "systemd.netname", "ifcfg", "NetworkManager", "debian.interfaces", "wicked", "ignition"}

	for _, overlay := range append(runtimeOverlays, systemOverlays...) {
		env.ImportDir(
			path.Join("var/lib/warewulf/overlays", overlay, "rootfs"),
			path.Join("../../../../../overlays", overlay, "rootfs"))
	}

	baseCmd.SetArgs([]string{"tn[1-999]"})
	buf := new(bytes.Buffer)
	baseCmd.SetOut(buf)
	baseCmd.SetErr(buf)
	wwlog.SetLogWriter(buf)
	for i := 0; i < b.N; i++ {
		err := baseCmd.Execute()
		if err != nil {
			b.Errorf("%s", err)
		}
	}
}
