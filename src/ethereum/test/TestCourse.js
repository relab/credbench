const { BN, expectEvent, constants, time, expectRevert } = require('@openzeppelin/test-helpers');
const { expect } = require('chai');

const Course = artifacts.require('CourseMock');

contract('Course', accounts => {
    const [teacher, evaluator, student, other, another] = accounts;
    let course = null;
    const exam1 = web3.utils.keccak256(web3.utils.toHex('cert1'));
    const exam2 = web3.utils.keccak256(web3.utils.toHex('cert2'));
    const exam3 = web3.utils.keccak256(web3.utils.toHex('cert3'));
    const courseDigest = web3.utils.keccak256(web3.utils.toHex('course1'));

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            course = await Course.new([teacher, evaluator], 2);
            (await course.isOwner(teacher)).should.equal(true);
            (await course.isOwner(evaluator)).should.equal(true);
            expect(await course.quorum()).to.be.bignumber.equal(new BN(2));
            ((await course.getStudents()).length).should.equal(0);
        });
    });

    describe('add student', () => {

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
        });

        it('should add a new student', async () => {
            const { logs } = await course.addStudent(student, { from: teacher });

            expectEvent.inLogs(logs, 'StudentAdded', {
                student: student,
                createdBy: teacher
            });

            (await course.isEnrolled(student)).should.equal(true);
            let students = await course.getStudents();
            expect(students).to.include.members([student]);
            (students.length).should.equal(1);

            let s = await course.getStudent(student);
            expect(s.index).to.be.bignumber.equal(new BN(0));
            (s.enrolled).should.equal(true);
        });

        it('should correctly increment the index when adding new students', async () => {
            await course.addStudent(student, { from: teacher });

            let s1 = await course.getStudent(student);
            expect(s1.index).to.be.bignumber.equal(new BN(0));
            (s1.enrolled).should.equal(true);

            await course.addStudent(other, { from: teacher });

            let s2 = await course.getStudent(other);
            expect(s2.index).to.be.bignumber.equal(new BN(1));
            (s2.enrolled).should.equal(true);
        });

        it('new students cannot be owner', async () => {
            await expectRevert(
                course.addStudent(teacher, { from: evaluator }),
                "Course/student cannot be owner"
            );
        });

        it('only owner should be able to add a student', async () => {
            await expectRevert(
                course.addStudent(student, { from: other }),
                "Owners/sender is not an owner"
            );
        });

        it('should not add a student twice', async () => {
            await course.addStudent(student, { from: teacher });
            await expectRevert(
                course.addStudent(student, { from: evaluator }),
                "Course/student already registered"
            );
        });

        it('should verify if a student is enrolled in the course', async () => {
            (await course.isEnrolled(student)).should.equal(false);
            await course.addStudent(student, { from: teacher });
            (await course.isEnrolled(student)).should.equal(true);
        });

        it('should not allows add zero address', async () => {
            await expectRevert(
                course.addStudent(constants.ZERO_ADDRESS, { from: teacher }),
                "Course/zero address given"
            );
        });
    });

    describe('remove student', () => {

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
            await course.addStudent(student, { from: teacher });
            await course.addStudent(other, { from: teacher });
            await course.addStudent(another, { from: teacher });
        });

        it('only owner should be able to remove an enrolled student', async () => {
            await expectRevert(
                course.removeStudent(student, { from: other }),
                'Owners/sender is not an owner'
            );
        });

        it('should allow an owner to remove an enrolled student', async () => {
            const { logs } = await course.removeStudent(student, { from: evaluator });

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                createdBy: evaluator
            });

            (await course.isEnrolled(student)).should.equal(false);
        });

        it('should correctly decrement the index when removing students', async () => {
            let students = await course.getStudents();
            (students.length).should.equal(3);

            let s2 = await course.getStudent(student);
            expect(s2.index).to.be.bignumber.equal(new BN(0));
            (s2.enrolled).should.equal(true);
            let s3 = await course.getStudent(another);
            expect(s3.index).to.be.bignumber.equal(new BN(2));
            (s3.enrolled).should.equal(true);

            await course.removeStudent(student, { from: teacher });

            students = await course.getStudents();
            (students.length).should.equal(2);
            expect(students).to.include.members([other, another]);
            // the last student should had being reindexed
            s3 = await course.getStudent(another);
            expect(s3.index).to.be.bignumber.equal(new BN(0));
            (s3.enrolled).should.equal(true);
        });

        it('should revert if try to remove an unregistered student', async () => {
            await expectRevert(
                course.removeStudent(evaluator, { from: teacher }),
                "Course/student not registered"
            );
        });

        it('should allow a student to renounce the course', async () => {
            const { logs } = await course.renounceCourse({ from: student });

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                createdBy: student
            });

            (await course.isEnrolled(student)).should.equal(false);
        });
    });

    describe('certification', () => {

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
            await course.enrollStudents([student, other]);
        });

        it('should issue a credential for a enrolled student', async () => {
            await course.registerExam(student, exam1, { from: teacher });
        });

        it('should not issue a credential for a non-enrolled address', async () => {
            await expectRevert(
                course.registerExam(another, exam1, { from: teacher }),
                'Course/student not registered'
            );
        });

        it('only owner should be allowed to issue a credential', async () => {
            await expectRevert(
                course.registerExam(student, exam1, { from: other }),
                'Owners/sender is not an owner'
            );
        });
    });

    describe('aggregation', () => {
        const digests = [exam1, exam2, exam3, courseDigest];
        const expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', digests));

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
            await course.enrollStudents([student, other]);

            for (d of digests) {
                await course.registerExam(student, d, { from: teacher });
                await course.registerExam(student, d, { from: evaluator });
                await course.approveCredential(d, { from: student });
                await time.increase(time.duration.seconds(1));
            }
        });

        it('should not perform aggregation for unregistered students', async () => {
            await expectRevert(
                course.aggregateCredentials(another, digests, { from: teacher }),
                'Issuer/there are no credentials'
            );
        });

        it('should aggregate the course certificates of a subject', async () => {
            await course.aggregateCredentials(student, digests, { from: teacher });

            (await course.hasRoot(student)).should.equal(true);
            let root = await course.getRoot(student);

            let expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', digests));
            (root).should.equal(expected);
        });
    });
});
